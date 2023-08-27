package server

import (
	"bytes"
	"halligalli/env"
	"io"
	"net/http"
)

func HttpGet(endpoint string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", env.Url(endpoint), nil)
	if err != nil {
		return nil, err
	}

	token := env.GetContext().Token
	req.Header.Set("Authorization", token.GetString())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return body, nil
}

func HttpPost(endpoint string, reqBody []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", env.Url(endpoint), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	token := env.GetContext().Token
	req.Header.Set("Authorization", token.GetString())
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
