package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"ToDoInfo/internal/log"
)

func GetResponseError(data []byte) error {
	type errorInfo struct {
		Code string `json:"code"`
	}
	type errorResponse struct {
		Error errorInfo `json:"error"`
	}

	errInfo := errorResponse{}
	err := json.Unmarshal(data, &errInfo)
	if err != nil {
		return err
	}
	if errInfo.Error.Code == "" {
		return nil
	}

	return &ResponseError{errInfo.Error.Code}
}

func GetAuthError(data []byte) error {
	type errorInfo struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	errInfo := errorInfo{}
	err := json.Unmarshal(data, &errInfo)
	if err != nil {
		return err
	}
	if errInfo.Error == "" {
		return nil
	}

	return &ResponseError{errInfo.ErrorDescription}
}

func Post(requestUrl string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetRequest(requestUrl string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	var bearer = "Bearer " + token
	req.Header.Add("Authorization", bearer)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
