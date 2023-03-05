package discord

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/query"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

// OnMessage handle message event
type OnMessage func(context.Context, InteractionRequest) (InteractionResponse, func(context.Context) InteractionResponse)

var discordRequest = request.New().URL("https://discord.com/api/v8")

// App of package
type App struct {
	tracer        trace.Tracer
	handler       OnMessage
	applicationID string
	clientID      string
	clientSecret  string
	website       string
	publicKey     []byte
}

// Config of package
type Config struct {
	applicationID *string
	publicKey     *string
	clientID      *string
	clientSecret  *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		applicationID: flags.String(fs, prefix, "discord", "ApplicationID", "Application ID", "", overrides),
		publicKey:     flags.String(fs, prefix, "discord", "PublicKey", "Public Key", "", overrides),
		clientID:      flags.String(fs, prefix, "discord", "ClientID", "Client ID", "", overrides),
		clientSecret:  flags.String(fs, prefix, "discord", "ClientSecret", "Client Secret", "", overrides),
	}
}

// New creates new App from Config
func New(config Config, website string, handler OnMessage, tracer trace.Tracer) (App, error) {
	publicKeyStr := *config.publicKey
	if len(publicKeyStr) == 0 {
		return App{}, nil
	}

	publicKey, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		return App{}, fmt.Errorf("decode public key string: %w", err)
	}

	return App{
		tracer:        tracer,
		applicationID: *config.applicationID,
		publicKey:     publicKey,
		clientID:      *config.clientID,
		clientSecret:  *config.clientSecret,
		website:       website,
		handler:       handler,
	}, nil
}

// Handler for request. Should be use with net/http
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth" {
			a.handleOauth(w, r)
			return
		}

		if !a.checkSignature(r) {
			httperror.Unauthorized(w, errors.New("invalid signature"))
			return
		}

		if query.IsRoot(r) && r.Method == http.MethodPost {
			a.handleWebhook(w, r)
			return
		}

		httperror.NotFound(w)
	})
}

func (a App) checkSignature(r *http.Request) bool {
	sig, err := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
	if err != nil {
		logger.Warn("decode signature string: %s", err)
		return false
	}

	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		logger.Warn("length of signature is invalid: %d", len(sig))
		return false
	}

	body, err := request.ReadBodyRequest(r)
	if err != nil {
		logger.Warn("read request body: %s", err)
		return false
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var msg bytes.Buffer
	msg.WriteString(r.Header.Get("X-Signature-Timestamp"))
	msg.Write(body)

	return ed25519.Verify(a.publicKey, msg.Bytes(), sig)
}

func (a App) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var (
		message InteractionRequest
		err     error
	)

	ctx, end := tracer.StartSpan(r.Context(), a.tracer, "webhook")
	defer end(&err)

	if err = httpjson.Parse(r, &message); err != nil {
		httperror.BadRequest(w, err)
		return
	}

	if message.Type == pingInteraction {
		httpjson.Write(w, http.StatusOK, InteractionResponse{Type: pongCallback})
		return
	}

	response, asyncFn := a.handler(ctx, message)
	httpjson.Write(w, http.StatusOK, response)

	if asyncFn != nil {
		go func(ctx context.Context) {
			var err error

			ctx, end := tracer.StartSpan(ctx, a.tracer, "async_webhook")
			defer end(&err)

			deferredResponse := asyncFn(ctx)

			resp, err := a.send(ctx, http.MethodPatch, fmt.Sprintf("/webhooks/%s/%s/messages/@original", a.applicationID, message.Token), deferredResponse.Data)
			if err != nil {
				logger.Error("send async response: %s", err)
				return
			}

			if err = request.DiscardBody(resp.Body); err != nil {
				logger.Error("discard async body: %s", err)
			}
		}(cntxt.WithoutDeadline(ctx))
	}
}

func (a App) send(ctx context.Context, method, path string, data InteractionDataResponse) (resp *http.Response, err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "send")
	defer end(&err)

	req := discordRequest.Method(method).Path(path)

	if len(data.Attachments) > 0 {
		return req.Multipart(ctx, writeMultipart(data))
	}

	return req.StreamJSON(ctx, data)
}

func writeMultipart(data InteractionDataResponse) func(*multipart.Writer) error {
	return func(mw *multipart.Writer) error {
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", `form-data; name="payload_json"`)
		header.Set("Content-Type", "application/json")
		partWriter, err := mw.CreatePart(header)
		if err != nil {
			return fmt.Errorf("create payload part: %w", err)
		}

		if err = json.NewEncoder(partWriter).Encode(data); err != nil {
			return fmt.Errorf("encode payload part: %w", err)
		}

		for _, file := range data.Attachments {
			file.Ephemeral = data.Flags&EphemeralMessage != 0

			if err = addAttachment(mw, file); err != nil {
				return err
			}
		}

		return nil
	}
}

func addAttachment(mw *multipart.Writer, file Attachment) error {
	partWriter, err := mw.CreateFormFile(fmt.Sprintf("files[%d]", file.ID), file.Filename)
	if err != nil {
		return fmt.Errorf("create file part: %w", err)
	}

	var fileReader io.ReadCloser
	fileReader, err = os.Open(file.filepath)
	if err != nil {
		return fmt.Errorf("open file part: %w", err)
	}

	defer func() {
		if closeErr := fileReader.Close(); closeErr != nil {
			logger.Error("close file part: %s", closeErr)
		}
	}()

	if _, err = io.Copy(partWriter, fileReader); err != nil {
		return fmt.Errorf("copy file part: %w", err)
	}

	return nil
}
