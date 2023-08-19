package main

import (
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	token := LoadFromConfig("./config.yaml")

	url, err := GetWebSocketUrl(token)
	if err != nil {
		log.Panicln("ERROR unable to get websocket url", err)
	}

	// relay the interrupt message from the system to the interrupt channel
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Printf("connecting to %s", url)
	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Panicln("ERROR connecting to :", err)
	}
	defer func(conn *websocket.Conn) {
		if err := conn.Close(); err != nil {
			log.Panicln("close:", err)
		}
	}(connection)

	ticker := time.NewTicker(time.Duration(DefaultHeartbeatIntervalMillis) * time.Millisecond)
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()

	lastMessageId := DefaultLastMessageId

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Println("ERROR reading message:", err)
				continue
			}
			log.Printf("receive: %s", message)
			raw, err := GetOpType(message)
			if err != nil {
				log.Println("ERROR getting operation type:", err)
				continue
			}
			if raw.MessageId != DefaultLastMessageId {
				lastMessageId = raw.MessageId
			}

			switch raw.Op {
			case Hello:
				if ticker, err = HandleHelloResponse(connection, raw.Body, token); err != nil {
					continue
				}
			case HeartbeatAck:
				log.Printf("heartbeat acknowledged")
			case Dispatch:
				log.Printf("intent: %s", raw.Intent)
				switch raw.Intent {
				case Ready:
					if err := HandleReadyMessageResponse(raw.Body); err != nil {
						continue
					}
				}
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			var heartbeatReq HeartbeatBody
			if lastMessageId != DefaultLastMessageId {
				heartbeatReq.LastMessageId = strconv.Itoa(lastMessageId)
			}
			request, err := BuildRequest(Heartbeat, heartbeatReq).GetString()
			if err != nil {
				log.Println("ERROR building request:", err)
				continue
			}
			err = connection.WriteMessage(websocket.TextMessage, request)
			if err != nil {
				log.Println("ERROR sending message:", err)
				continue
			}
			log.Printf("heartbeat, last message id: %d", lastMessageId)
		case <-interrupt:
			log.Println("interrupted by user event")
			err := connection.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("ERROR sending message:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
