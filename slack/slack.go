package slack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

// CommandHandler for handling when user send a slash command
type CommandHandler func(context.Context, SlashPayload) Response

// InteractHandler for handling when user interact with a button
type InteractHandler func(context.Context, InteractivePayload) Response

type Config struct {
	ClientID      string
	ClientSecret  string
	SigningSecret string
}

type Service struct {
	tracer     trace.Tracer
	onCommand  CommandHandler
	onInteract InteractHandler

	clientID      string
	clientSecret  string
	signingSecret []byte
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	var config Config

	flags.New("ClientID", "ClientID").Prefix(prefix).DocPrefix("slack").StringVar(fs, &config.ClientID, "", overrides)
	flags.New("ClientSecret", "ClientSecret").Prefix(prefix).DocPrefix("slack").StringVar(fs, &config.ClientSecret, "", overrides)
	flags.New("SigningSecret", "Signing secret").Prefix(prefix).DocPrefix("slack").StringVar(fs, &config.SigningSecret, "", overrides)

	return config
}

func New(config Config, command CommandHandler, interact InteractHandler, tracerProvider trace.TracerProvider) Service {
	app := Service{
		clientID:      config.ClientID,
		clientSecret:  config.ClientSecret,
		signingSecret: []byte(config.SigningSecret),

		onCommand:  command,
		onInteract: interact,
	}

	if tracerProvider != nil {
		app.tracer = tracerProvider.Tracer("slack")
	}

	return app
}

func (a Service) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth" {
			a.handleOauth(w, r)
			return
		}

		if !a.checkSignature(r) {
			httperror.Unauthorized(w, errors.New("invalid signature"))
			return
		}

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusOK)

		case http.MethodPost:
			if r.URL.Path == "/interactive" {
				a.handleInteract(w, r)
			} else {
				payload := SlashPayload{
					ChannelID:   r.FormValue("channel_id"),
					Command:     strings.TrimPrefix(r.FormValue("command"), "/"),
					ResponseURL: r.FormValue("response_url"),
					Text:        r.FormValue("text"),
					Token:       r.FormValue("token"),
					UserID:      r.FormValue("user_id"),
				}

				httpjson.Write(w, http.StatusOK, a.onCommand(r.Context(), payload))
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (a Service) checkSignature(r *http.Request) bool {
	tsValue, err := strconv.ParseInt(r.Header.Get("X-Slack-Request-Timestamp"), 10, 64)
	if err != nil {
		slog.Error("parse timestamp", "err", err)
		return false
	}

	if time.Unix(tsValue, 0).Before(time.Now().Add(time.Minute * -5)) {
		slog.Warn("timestamp is from 5 minutes ago")
		return false
	}

	body, err := request.ReadBodyRequest(r)
	if err != nil {
		slog.Warn("read request body", "err", err)
		return false
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	slackSignature := r.Header.Get("X-Slack-Signature")
	signatureValue := []byte(fmt.Sprintf("v0:%d:%s", tsValue, body))

	sig := hmac.New(sha256.New, a.signingSecret)
	sig.Write(signatureValue)
	ownSignature := fmt.Sprintf("v0=%s", hex.EncodeToString(sig.Sum(nil)))

	if hmac.Equal([]byte(slackSignature), []byte(ownSignature)) {
		return true
	}

	slog.Error("signature mismatch from slack's one", "slack_signature", slackSignature)
	return false
}

func (a Service) handleInteract(w http.ResponseWriter, r *http.Request) {
	var (
		payload InteractivePayload
		err     error
	)

	ctx, end := telemetry.StartSpan(r.Context(), a.tracer, "interact")
	defer end(&err)

	if err := json.Unmarshal([]byte(r.FormValue("payload")), &payload); err != nil {
		httpjson.Write(w, http.StatusOK, NewEphemeralMessage(fmt.Sprintf("cannot unmarshall payload: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)

	go func(ctx context.Context) {
		var err error

		ctx, end := telemetry.StartSpan(ctx, a.tracer, "async_intereact")
		defer end(&err)

		slackResponse := a.onInteract(ctx, payload)

		resp, err := request.Post(payload.ResponseURL).StreamJSON(ctx, slackResponse)
		if err != nil {
			slog.Error("send interact on response_url", "err", err)
		} else if discardErr := request.DiscardBody(resp.Body); discardErr != nil {
			slog.Error("discard interact body on response_url", "err", err)
		}
	}(cntxt.WithoutDeadline(ctx))
}
