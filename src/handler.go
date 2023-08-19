package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func HandleHelloResponse(body json.RawMessage) (*time.Ticker, error) {
	helloResp, err := ParseHelloResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return nil, err
	}
	// set heartbeat ticker
	ticker := time.NewTicker(time.Duration(helloResp.HeartbeatInterval) * time.Millisecond)

	// send identify message
	identifyReq := IdentifyBody{
		Token:      GetContext().Token.GetString(),
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
	err = GetContext().Connection.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		log.Println("ERROR sending message:", err)
		return nil, err
	}
	return ticker, nil
}

func HandleReadyResponse(body json.RawMessage) error {
	readyResp, err := ParseReadyResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return err
	}
	GetContext().User = readyResp.User
	log.Printf("logged in as user %s %s, session id: %s",
		readyResp.User.UserName, readyResp.User.Id, readyResp.SessionId)
	return nil
}

func HandleMessageCreateResponse(body json.RawMessage) error {
	messageCreateBody, err := ParseMessageCreateResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return err
	}

	hasBotMentioned := false
	for _, mention := range messageCreateBody.Mentions {
		if mention.Id == GetContext().User.Id {
			hasBotMentioned = true
		}
	}

	if hasBotMentioned == false {
		return nil
	}

	url := fmt.Sprintf("/channels/%s/messages", messageCreateBody.ChannelId)
	reqBody := MessageSendBody{
		Content:        "喵~",
		ReplyMessageId: messageCreateBody.Id,
	}
	reqBodyRaw, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	log.Printf("debug log: %s", reqBodyRaw)

	respRaw, err := HttpPost(url, reqBodyRaw)
	if err != nil {
		return err
	}

	log.Printf("sending message response: %s", string(respRaw))
	return nil
}
