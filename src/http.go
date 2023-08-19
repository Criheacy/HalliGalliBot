package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func GetWebSocketUrl(token *Token) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", Url("/gateway"), nil)

	req.Header.Set("Authorization", token.GetString())
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	err = res.Body.Close()
	if err != nil {
		return "", err
	}

	var bodyParam map[string]string
	err = json.Unmarshal(body, &bodyParam)
	if err != nil {
		return "", err
	}

	return bodyParam["url"], nil
}
