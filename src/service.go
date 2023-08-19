package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

func GetWebsocketUrl() (string, error) {
	bodyRaw, err := HttpGet("/gateway")
	if err != nil {
		log.Println("ERROR unable to get websocket url", err)
		return "", err
	}
	var gatewayResp GatewayBody
	if err = json.Unmarshal(bodyRaw, &gatewayResp); err != nil {
		log.Println("ERROR parsing gateway response", err)
		return "", err
	}
	return gatewayResp.Url, nil
}

func ConnectToWebsocketServer() error {
	url, err := GetWebsocketUrl()
	if err != nil {
		return err
	}
	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("ERROR connecting to :", err)
		return err
	}
	GetContext().Connection = connection
	return nil
}
