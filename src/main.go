package main

import (
	"github.com/gorilla/websocket"
	"halligalli/auth"
	"halligalli/env"
	"halligalli/game"
	"halligalli/model"
	"halligalli/server"
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

	err := game.LoadAssets()
	if err != nil {
		log.Panicln("ERROR loading assets", err)
	}

	err = auth.LoadTokenFromConfig()
	if err != nil {
		log.Panicln("ERROR load token from config", err)
	}

	err = server.ConnectToWebsocketServer()
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
	}(env.GetContext().Connection)

	ticker := time.NewTicker(time.Duration(model.DefaultHeartbeatIntervalMillis) * time.Millisecond)
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()

	lastMessageId := model.DefaultLastMessageId

	eventChannel := make(chan game.Event, 8)
	messageChannel := make(chan game.Message, 8)

	go game.MainLoop(eventChannel, messageChannel)
	go server.HandleGameMessage(messageChannel)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := env.GetContext().Connection.ReadMessage()
			if err != nil {
				log.Println("ERROR reading message:", err)
				continue
			}
			log.Printf("receive: %s", message)
			raw, err := model.GetOpType(message)
			if err != nil {
				log.Println("ERROR getting operation type:", err)
				continue
			}
			if raw.MessageId != model.DefaultLastMessageId {
				lastMessageId = raw.MessageId
			}

			switch raw.Op {
			case model.Hello:
				if ticker, err = server.HandleHelloResponse(raw.Body); err != nil {
					continue
				}
			case model.HeartbeatAck:
				log.Printf("heartbeat acknowledged")
			case model.Dispatch:
				switch raw.Intent {
				case model.Ready:
					if err := server.HandleReadyResponse(raw.Body); err != nil {
						log.Println("ERROR handle ready response", err)
						continue
					}
				case model.MessageCreate:
					if err := server.HandleMessageCreateResponse(raw.Body, eventChannel); err != nil {
						log.Println("ERROR handle message", err)
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
			var heartbeatReq model.HeartbeatBody
			if lastMessageId != model.DefaultLastMessageId {
				heartbeatReq.LastMessageId = strconv.Itoa(lastMessageId)
			}
			request, err := model.BuildRequest(model.Heartbeat, heartbeatReq).GetString()
			if err != nil {
				log.Println("ERROR building request:", err)
				continue
			}
			err = env.GetContext().Connection.WriteMessage(websocket.TextMessage, request)
			if err != nil {
				log.Println("ERROR sending message:", err)
				continue
			}
			log.Printf("heartbeat, last message id: %d", lastMessageId)
		case <-interrupt:
			log.Println("interrupted by user event")
			err := env.GetContext().Connection.WriteMessage(websocket.CloseMessage,
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
