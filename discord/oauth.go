package discord

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func (s Service) handleOauth(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	params.Set("code", r.URL.Query().Get("code"))
	params.Set("client_id", s.clientID)
	params.Set("client_secret", s.clientSecret)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", s.website)

	resp, err := discordRequest.Path("/oauth2/token").Method(http.MethodPost).Form(r.Context(), params)
	if err != nil {
		httperror.InternalServerError(r.Context(), w, fmt.Errorf("confirm oauth request: %w", err))
		return
	}

	if err := request.DiscardBody(resp.Body); err != nil {
		httperror.InternalServerError(r.Context(), w, fmt.Errorf("discard body: %w", err))
		return
	}

	http.Redirect(w, r, s.website, http.StatusFound)
}
