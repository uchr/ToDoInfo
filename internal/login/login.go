package login

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"ToDoInfo/internal/httpclient"
)

type authData struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
}

type tokenData struct {
	AccessToken string `json:"access_token"`
}

const (
	baseRequestUrl = "https://login.microsoftonline.com/consumers/oauth2/v2.0"
)

func authRequest(clientId string) (*authData, error) {
	const authRequestUrl = "/devicecode"

	values := url.Values{}
	values.Add("client_id", clientId)
	values.Add("scope", "User.Read Tasks.ReadWrite")

	responseBody, err := httpclient.PostRequest(baseRequestUrl+authRequestUrl, values)
	if err != nil {
		return nil, err
	}

	err = httpclient.GetResponseError(responseBody)
	if err != nil {
		return nil, err
	}

	auth := authData{}
	err = json.Unmarshal(responseBody, &auth)
	if err != nil {
		return nil, err
	}

	return &auth, nil
}

func tokenRequest(clientId string, deviceCode string) (string, error) {
	const tokenRequestUrl = "/token"

	values := url.Values{}
	values.Add("client_id", clientId)
	values.Add("code", deviceCode)
	values.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	responseBody, err := httpclient.PostRequest(baseRequestUrl+tokenRequestUrl, values)
	if err != nil {
		return "", err
	}

	err = httpclient.GetResponseError(responseBody)
	if err != nil {
		return "", err
	}

	token := tokenData{}
	err = json.Unmarshal(responseBody, &token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func checkTokenExpiration(token string) (bool, error) {
	const checkUrl = "https://graph.microsoft.com/v1.0/me"
	if token == "" {
		return true, nil
	}

	responseBody, err := httpclient.GetRequest(checkUrl, token)
	if err != nil {
		return false, err
	}

	err = httpclient.GetResponseError(responseBody)
	if err != nil {
		responseErr := &httpclient.ResponseError{}
		if errors.As(err, &responseErr); responseErr.Code == httpclient.InvalidAuthenticationTokenCode {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func loadToken() (string, error) {
	data, err := os.ReadFile(".cache")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	return string(data), nil
}

func storeToken(token string) error {
	err := os.WriteFile(".cache", []byte(token), 0644)
	if err != nil {
		return err
	}
	return nil
}

func Login(clientId string) (string, error) {
	token, err := loadToken()
	if err != nil {
		return "", err
	}

	isExpired, err := checkTokenExpiration(token)
	if err != nil {
		return "", err
	}
	if !isExpired {
		return token, nil
	}

	auth, err := authRequest(clientId)
	if err != nil {
		return "", err
	}

	fmt.Println(fmt.Sprintf("To sign in, use a web browser to open the page %s and enter the code %s to authenticate.",
		auth.VerificationUri, auth.UserCode))
	fmt.Println("Press 'Enter' after login to continue...")
	_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return "", err
	}

	token, err = tokenRequest(clientId, auth.DeviceCode)
	if err != nil {
		return "", err
	}

	err = storeToken(token)
	if err != nil {
		return "", err
	}

	return token, nil
}
