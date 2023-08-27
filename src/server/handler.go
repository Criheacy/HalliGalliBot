package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"halligalli/env"
	"halligalli/game"
	"halligalli/model"
	"log"
	"strings"
	"time"
)

func HandleHelloResponse(body json.RawMessage) (*time.Ticker, error) {
	helloResp, err := model.ParseHelloResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return nil, err
	}
	// set heartbeat ticker
	ticker := time.NewTicker(time.Duration(helloResp.HeartbeatInterval) * time.Millisecond)

	// send identify message
	identifyReq := model.IdentifyBody{
		Token:      env.GetContext().Token.GetString(),
		Intents:    model.Intents,
		Shard:      [2]int{0, 1},
		Properties: map[string]string{},
	}
	req, err := model.BuildRequest(model.Identify, identifyReq).GetString()
	if err != nil {
		log.Println("ERROR building request:", err)
		return nil, err
	}
	log.Printf("authenticate: %s", req)
	err = env.GetContext().Connection.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		log.Println("ERROR sending message:", err)
		return nil, err
	}
	return ticker, nil
}

func HandleReadyResponse(body json.RawMessage) error {
	readyResp, err := model.ParseReadyResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return err
	}
	env.GetContext().User = readyResp.User
	log.Printf("logged in as user %s %s, session id: %s",
		readyResp.User.UserName, readyResp.User.Id, readyResp.SessionId)
	return nil
}

func HandleMessageCreateResponse(body json.RawMessage, eventChannel chan game.Event) error {
	messageCreateBody, err := model.ParseMessageCreateResponseBody(body)
	if err != nil {
		log.Println("ERROR parsing response body:", err)
		return err
	}

	hasBotMentioned := false
	for _, mention := range messageCreateBody.Mentions {
		if mention.Id == env.GetContext().User.Id {
			hasBotMentioned = true
		}
	}

	if hasBotMentioned == false {
		return nil
	}

	env.GetContext().ReplyMessageId = messageCreateBody.Id
	if strings.Contains(messageCreateBody.Content, "game") {
		eventChannel <- game.Event{
			EventType: game.Initiate,
			ChannelId: messageCreateBody.ChannelId,
			Param:     nil,
		}
	} else if strings.Contains(messageCreateBody.Content, "start") {
		eventChannel <- game.Event{
			EventType: game.Start,
			ChannelId: messageCreateBody.ChannelId,
			Param:     nil,
		}
	} else if strings.Contains(messageCreateBody.Content, "continue") {
		eventChannel <- game.Event{
			EventType: game.Continue,
			ChannelId: messageCreateBody.ChannelId,
			Param:     nil,
		}
	} else if strings.Contains(messageCreateBody.Content, "stop") {
		eventChannel <- game.Event{
			EventType: game.Terminate,
			ChannelId: messageCreateBody.ChannelId,
			Param:     nil,
		}
	} else {
		eventChannel <- game.Event{
			EventType: game.RingTheBell,
			ChannelId: messageCreateBody.ChannelId,
			Param:     messageCreateBody.Author,
		}
	}
	return nil
}

func HandleGameMessage(messageChannel chan game.Message) {
	for {
		select {
		case message := <-messageChannel:
			var messageBody model.MessageSendBody
			switch message.MessageType {
			case game.ShowGameRule:
				messageBody = model.MessageSendBody{
					Content: "欢迎来到 HalliGalli 小游戏！\n" +
						"接下来我会依次翻开带有水果或动物图案的牌，如果在翻开的最后 5 张牌中有 5 个相同的水果或者含有动物牌，" +
						"请立即发送一条 @我 的消息表示您按响了铃铛！\n" +
						"第一个按响铃铛的玩家会赢下本轮，并由我重新发牌。小心不要按错了哦！\n" +
						"准备好了吗？请 @我 发送 \"start\" 来开始游戏！",
				}
			case game.CardRevealed:
				card := message.Param.(game.Card)
				messageBody = model.MessageSendBody{
					ImageUrl: card.Image,
				}
			case game.PlayerWin:
				roundStatus := message.Param.(game.RoundStatus)
				mentionPlayer := fmt.Sprintf("<@!%s>", roundStatus.Player.Id)
				var reason string
				if roundStatus.FruitName != "" {
					reason = fmt.Sprintf(" 5 个%s", roundStatus.FruitName)
				} else if roundStatus.AnimalName != "" {
					reason = fmt.Sprintf("%s", roundStatus.AnimalName)
				}
				messageBody = model.MessageSendBody{
					Content: "恭喜" + mentionPlayer + "赢得了这一轮！\n" +
						"（最后五张牌中有" + reason + "）\n" +
						"准备好清空桌面！@我 发送 continue 开始新的一轮！",
				}
			case game.FakeRing:
				roundStatus := message.Param.(game.RoundStatus)
				atPlayer := fmt.Sprintf("<@!%s>", roundStatus.Player.Id)
				messageBody = model.MessageSendBody{
					Content: atPlayer + "非常遗憾！桌面上并不满足按铃的条件！\n" +
						"不要灰心丧气！重整旗鼓，@我 发送 continue 继续游戏！",
				}
			case game.Terminated:
				messageBody = model.MessageSendBody{
					Content: "游戏告一段落啦！想要再来一局，请随时 @我 发送 game 哦！",
				}
			}
			messageBody.ReplyMessageId = env.GetContext().ReplyMessageId
			err := SendMessage(message.ChannelId, &messageBody)
			if err != nil {
				log.Println("ERROR sending message", err)
				continue
			}
		}
	}
}
