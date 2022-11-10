package login

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/uchr/ToDoInfo/internal/config"
	"github.com/uchr/ToDoInfo/internal/httpclient"
)

type tokenData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

const (
	baseRequestUrl  = "https://login.microsoftonline.com/consumers/oauth2/v2.0"
	authRequestUrl  = "/authorize"
	tokenRequestUrl = "/token"
)

func GetAuthRequest(cfg config.Config) string {
	values := url.Values{}
	values.Add("client_id", cfg.ClientId)
	values.Add("redirect_uri", cfg.RedirectURI)
	values.Add("response_type", "code")
	values.Add("response_mode", "query")
	values.Add("scope", "User.Read Tasks.ReadWrite")
	return baseRequestUrl + authRequestUrl + "?" + values.Encode()
}

func Auth(ctx context.Context, cfg config.Config, code string) (string, time.Duration, error) {
	values := url.Values{}
	values.Add("client_id", cfg.ClientId)
	values.Add("client_secret", cfg.ClientSecret)
	values.Add("code", code)
	values.Add("redirect_uri", cfg.RedirectURI)
	values.Add("grant_type", "authorization_code")

	body, err := httpclient.Post(ctx, baseRequestUrl+tokenRequestUrl, values)
	if err != nil {
		return "", 0, err
	}

	err = httpclient.GetAuthError(body)
	if err != nil {
		return "", 0, err
	}

	token := tokenData{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", 0, err
	}

	return token.AccessToken, time.Duration(token.ExpiresIn * int(time.Second)), nil
}
