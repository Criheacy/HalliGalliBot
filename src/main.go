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
	// relay the interrupt message from the system to the interrupt channel
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	err := LoadAssets()
	if err != nil {
		log.Panicln("ERROR loading assets", err)
	}
	log.Printf("assets: %+v", GetContext().Asset)

	err = LoadTokenFromConfig()
	if err != nil {
		log.Panicln("ERROR load token from config", err)
	}

	err = ConnectToWebsocketServer()
	if err != nil {
		log.Panicln("ERROR connect to websocket server", err)
	}
	defer func(conn *websocket.Conn) {
		if conn == nil {
			return
		}
		if err := conn.Close(); err != nil {
			log.Panicln("close:", err)
		}
	}(GetContext().Connection)

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
			_, message, err := GetContext().Connection.ReadMessage()
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
				if ticker, err = HandleHelloResponse(raw.Body); err != nil {
					continue
				}
			case HeartbeatAck:
				log.Printf("heartbeat acknowledged")
			case Dispatch:
				log.Printf("intent: %s", raw.Intent)
				switch raw.Intent {
				case Ready:
					if err := HandleReadyResponse(raw.Body); err != nil {
						continue
					}
				case MessageCreate:
					if err := HandleMessageCreateResponse(raw.Body); err != nil {
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
			err = GetContext().Connection.WriteMessage(websocket.TextMessage, request)
			if err != nil {
				log.Println("ERROR sending message:", err)
				continue
			}
			log.Printf("heartbeat, last message id: %d", lastMessageId)
		case <-interrupt:
			log.Println("interrupted by user event")
			err := GetContext().Connection.WriteMessage(websocket.CloseMessage,
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
