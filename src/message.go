package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func SendMessage(body MessageSendBody) error {
	url := fmt.Sprintf("/channels/%s/messages", GetContext().ChannelId)
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	log.Printf("send: %s", bodyRaw)

	respRaw, err := HttpPost(url, bodyRaw)
	if err != nil {
		return err
	}

	log.Printf("sending message response: %s", string(respRaw))
	return nil
}
