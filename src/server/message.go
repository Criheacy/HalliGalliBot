package server

import (
	"encoding/json"
	"fmt"
	"halligalli/model"
	"log"
)

func SendMessage(channelId string, body *model.MessageSendBody) error {
	url := fmt.Sprintf("/channels/%s/messages", channelId)
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
