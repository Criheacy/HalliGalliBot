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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Printf("connecting to %s", url)

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Panicln("ERROR connecting to :", err)
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
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
				helloResp, err := ParseHelloResponseBody(raw.Body)
				if err != nil {
					log.Println("ERROR parsing response body:", err)
					continue
				}
				// set heartbeat ticker
				ticker = time.NewTicker(time.Duration(helloResp.HeartbeatInterval) * time.Millisecond)

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
					continue
				}
				log.Printf("authenticate: %s", req)
				err = connection.WriteMessage(websocket.TextMessage, req)
				if err != nil {
					log.Println("ERROR sending message:", err)
					continue
				}
			case HeartbeatAck:
				log.Printf("heartbeat acknowledged")
			case Dispatch:
				log.Printf("intent: %s", raw.Intent)
				switch raw.Intent {
				case Ready:
					readyResp, err := ParseReadyMessageResponseBody(raw.Body)
					if err != nil {
						log.Println("ERROR parsing response body:", err)
						return
					}
					log.Printf("logged in as user %s %s, session id: %s",
						readyResp.User.UserName, readyResp.User.Id, readyResp.SessionId)
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
