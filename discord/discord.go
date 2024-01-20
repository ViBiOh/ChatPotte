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
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/query"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

// OnMessage handle message event
type OnMessage func(context.Context, InteractionRequest) (InteractionResponse, func(context.Context) InteractionResponse)

var discordRequest = request.New().URL("https://discord.com/api/v8")

type Service struct {
	tracer        trace.Tracer
	handler       OnMessage
	applicationID string
	clientID      string
	clientSecret  string
	website       string
	publicKey     []byte
}

type Config struct {
	ApplicationID string
	PublicKey     string
	ClientID      string
	ClientSecret  string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("ApplicationID", "Application ID").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.ApplicationID, "", overrides)
	flags.New("PublicKey", "Public Key").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.PublicKey, "", overrides)
	flags.New("ClientID", "Client ID").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.ClientID, "", overrides)
	flags.New("ClientSecret", "Client Secret").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.ClientSecret, "", overrides)

	return &config
}

func New(config *Config, website string, handler OnMessage, tracerProvider trace.TracerProvider) (Service, error) {
	publicKeyStr := config.PublicKey
	if len(publicKeyStr) == 0 {
		return Service{}, nil
	}

	publicKey, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		return Service{}, fmt.Errorf("decode public key string: %w", err)
	}

	app := Service{
		applicationID: config.ApplicationID,
		publicKey:     publicKey,
		clientID:      config.ClientID,
		clientSecret:  config.ClientSecret,
		website:       website,
		handler:       handler,
	}

	if tracerProvider != nil {
		app.tracer = tracerProvider.Tracer("discord")
	}

	return app, nil
}

func (s Service) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth" {
			s.handleOauth(w, r)
			return
		}

		if !s.checkSignature(r) {
			httperror.Unauthorized(r.Context(), w, errors.New("invalid signature"))
			return
		}

		if query.IsRoot(r) && r.Method == http.MethodPost {
			s.handleWebhook(w, r)
			return
		}

		httperror.NotFound(r.Context(), w)
	})
}

func (s Service) checkSignature(r *http.Request) bool {
	sig, err := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
	if err != nil {
		slog.LogAttrs(r.Context(), slog.LevelWarn, "decode signature string", slog.Any("error", err))
		return false
	}

	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		slog.LogAttrs(r.Context(), slog.LevelWarn, "length of signature is invalid", slog.Int("length", len(sig)))
		return false
	}

	body, err := request.ReadBodyRequest(r)
	if err != nil {
		slog.LogAttrs(r.Context(), slog.LevelWarn, "read request body", slog.Any("error", err))
		return false
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var msg bytes.Buffer
	msg.WriteString(r.Header.Get("X-Signature-Timestamp"))
	msg.Write(body)

	return ed25519.Verify(s.publicKey, msg.Bytes(), sig)
}

func (s Service) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var (
		message InteractionRequest
		err     error
	)

	ctx, end := telemetry.StartSpan(r.Context(), s.tracer, "webhook")
	defer end(&err)

	if err = httpjson.Parse(r, &message); err != nil {
		httperror.BadRequest(r.Context(), w, err)
		return
	}

	if message.Type == pingInteraction {
		httpjson.Write(r.Context(), w, http.StatusOK, InteractionResponse{Type: pongCallback})
		return
	}

	response, asyncFn := s.handler(ctx, message)
	httpjson.Write(r.Context(), w, http.StatusOK, response)

	if asyncFn != nil {
		go func(ctx context.Context) {
			var err error

			ctx, end := telemetry.StartSpan(ctx, s.tracer, "async_webhook")
			defer end(&err)

			deferredResponse := asyncFn(ctx)

			resp, err := s.send(ctx, http.MethodPatch, fmt.Sprintf("/webhooks/%s/%s/messages/@original", s.applicationID, message.Token), deferredResponse.Data)
			if err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "send async response", slog.Any("error", err))
				return
			}

			if err = request.DiscardBody(resp.Body); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "discard async body", slog.Any("error", err))
			}
		}(cntxt.WithoutDeadline(ctx))
	}
}

func (s Service) send(ctx context.Context, method, path string, data InteractionDataResponse) (resp *http.Response, err error) {
	ctx, end := telemetry.StartSpan(ctx, s.tracer, "send")
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
			slog.LogAttrs(context.Background(), slog.LevelError, "close file part", slog.Any("error", closeErr))
		}
	}()

	if _, err = io.Copy(partWriter, fileReader); err != nil {
		return fmt.Errorf("copy file part: %w", err)
	}

	return nil
}
