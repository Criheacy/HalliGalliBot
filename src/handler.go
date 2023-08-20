package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strings"
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

func HandleMessageCreateResponse(body json.RawMessage, eventChannel chan GameEvent) error {
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

	if GetContext().ChannelId == "" {
		GetContext().ChannelId = messageCreateBody.ChannelId
	}

	if GetContext().ChannelId != messageCreateBody.ChannelId {
		// TODO: notify the user in other channel
		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(messageCreateBody.Content, "game") {
		eventChannel <- GameEvent{
			EventType: Initiate,
			Param:     nil,
		}
	} else if strings.Contains(messageCreateBody.Content, "start") {
		eventChannel <- GameEvent{
			EventType: Start,
			Param:     nil,
		}
	} else if strings.Contains(messageCreateBody.Content, "continue") {
		eventChannel <- GameEvent{
			EventType: Continue,
			Param:     nil,
		}
	} else if strings.Contains(messageCreateBody.Content, "stop") {
		eventChannel <- GameEvent{
			EventType: Terminate,
			Param:     nil,
		}
		GetContext().ChannelId = ""
	} else {
		eventChannel <- GameEvent{
			EventType: RingTheBell,
			Param:     messageCreateBody.Author,
		}
	}

	url := fmt.Sprintf("/channels/%s/messages", messageCreateBody.ChannelId)
	reqBody := MessageSendBody{
		ImageUrl: "https://p.sda1.dev/12/502e25ab8a81b2997bbe23c42f756740/Apple Watch 45mm - 1 _2_.png",
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

func HandleGameMessage(messageChannel chan GameMessage) {
	for {
		select {
		case message := <-messageChannel:
			var messageBody MessageSendBody
			switch message.MessageType {
			case ShowGameRule:
				messageBody = MessageSendBody{
					Content: "欢迎来到 HalliGalli 小游戏！\n" +
						"接下来我会依次翻开带有水果或动物图案的牌，如果在翻开的最后 5 张牌中有 5 个相同的水果或者含有动物牌，" +
						"请立即发送一条 @我 的消息表示您按响了铃铛！\n" +
						"第一个按响铃铛的玩家会赢下本轮，并由我重新发牌。小心不要按错了哦！\n" +
						"准备好了吗？请 @我 发送 \"start\" 来开始游戏！",
				}
			case CardRevealed:
				card := message.Param.(Card)
				messageBody = MessageSendBody{
					ImageUrl: card.Image,
				}
			case PlayerWin:
				roundStatus := message.Param.(RoundStatus)
				atPlayer := fmt.Sprintf("<@!%s>", roundStatus.Player.Id)
				var reason string
				if roundStatus.FruitVariant != -1 {
					name := GetContext().Asset.Meta.Fruits[roundStatus.FruitVariant].Name
					reason = fmt.Sprintf(" 5 个%s", name)
				} else if roundStatus.AnimalVariant != -1 {
					name := GetContext().Asset.Meta.Animals[roundStatus.AnimalVariant].Name
					reason = fmt.Sprintf("%s", name)
				}
				messageBody = MessageSendBody{
					Content: "恭喜" + atPlayer + roundStatus.Player.UserName + "赢得了这一轮！\n" +
						"（最后五张牌中有" + reason + "）\n" +
						"准备好清空桌面！@我 发送 continue 开始新的一轮！",
				}
			case FakeRing:
				roundStatus := message.Param.(RoundStatus)
				atPlayer := fmt.Sprintf("<@!%s>", roundStatus.Player.Id)
				messageBody = MessageSendBody{
					Content: atPlayer + "非常遗憾！桌面上并不满足按铃的条件！\n" +
						"不要灰心丧气！重整旗鼓，@我 发送 continue 继续游戏！",
				}
			case GameTerminated:
				messageBody = MessageSendBody{
					Content: "游戏告一段落啦！想要再来一句，请随时 @我 发送 game 哦！",
				}
			}
			err := SendMessage(messageBody)
			if err != nil {
				log.Println("ERROR sending message", err)
				continue
			}
		}
	}
}
