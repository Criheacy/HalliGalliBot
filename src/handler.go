package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func HandleHelloResponse(connection *websocket.Conn, body json.RawMessage, token Token) (*time.Ticker, error) {
	helloResp, err := ParseHelloResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return nil, err
	}
	// set heartbeat ticker
	ticker := time.NewTicker(time.Duration(helloResp.HeartbeatInterval) * time.Millisecond)

	// send identify message
	identifyReq := IdentifyBody{
		Token:      token.GetString(),
		Intents:    Intents,
		Shard:      [2]int{0, 1},
		Properties: map[string]string{},
	}
	req, err := BuildRequest(Identify, identifyReq).GetString()
	if err != nil {
		log.Println("ERROR building request:", err)
		return nil, err
	}
	log.Printf("authenticate: %s", req)
	err = connection.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		log.Println("ERROR sending message:", err)
		return nil, err
	}
	return ticker, nil
}

func HandleReadyMessageResponse(body json.RawMessage) error {
	readyResp, err := ParseReadyMessageResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return err
	}
	log.Printf("logged in as user %s %s, session id: %s",
		readyResp.User.UserName, readyResp.User.Id, readyResp.SessionId)
	return nil
}
