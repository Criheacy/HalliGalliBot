package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"halligalli/assets"
	"halligalli/auth"
	"halligalli/common"
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
		Token:      auth.GetTokenString(env.GetContext().Token),
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
	} else if strings.Contains(messageCreateBody.Content, "why") ||
		strings.Contains(messageCreateBody.Content, "debug") {
		eventChannel <- game.Event{
			EventType: game.Debug,
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
				card := message.Param.(common.Card)
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
					Content: fmt.Sprintf("恭喜%s赢得了这一轮！\n（最后五张牌中有%s）\n准备好清空桌面！@我 发送 continue 开始新的一轮！",
						mentionPlayer, reason),
				}
			case game.FakeRing:
				roundStatus := message.Param.(game.RoundStatus)
				atPlayer := fmt.Sprintf("<@!%s>", roundStatus.Player.Id)
				messageBody = model.MessageSendBody{
					Content: atPlayer + "非常遗憾！桌面上并不满足按铃的条件！\n不要灰心丧气！重整旗鼓，@我 发送 continue 继续游戏！",
				}
			case game.Terminated:
				messageBody = model.MessageSendBody{
					Content: "游戏告一段落啦！想要再来一局，请随时 @我 发送 game 哦！",
				}
			case game.ExplainWhy:
				validCards := message.Param.([]common.Card)
				messageBody = model.MessageSendBody{
					Content: BuildExplainMessage(validCards),
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

func BuildExplainMessage(validCards []common.Card) string {
	fruitCounter := make(map[int]int)
	animalCounter := make([]int, 0)
	var builder strings.Builder
	for index, card := range validCards {
		builder.WriteString(fmt.Sprintf("第%d张牌中", index))
		if card.Type == common.Animal {
			animalCounter = append(animalCounter, card.Variant)
			builder.WriteString(fmt.Sprintf("有一只%s", assets.GetAnimalNameByVariant(card.Variant)))
		} else if card.Type == common.Fruit {
			if len(card.Elements) == 0 {
				builder.WriteString("什么都没有")
			}
			firstFlag := true
			for _, fruit := range card.Elements {
				fruitCounter[fruit.Variant] += fruit.Number
				if firstFlag {
					firstFlag = false
					builder.WriteString("有")
				} else {
					builder.WriteString("、")
				}
				builder.WriteString(fmt.Sprintf("%d个%s", fruit.Number,
					assets.GetFruitNameByVariant(fruit.Variant)))
			}
		}
		builder.WriteString("\n")
	}
	builder.WriteString("总计")
	if len(fruitCounter) == 0 {
		builder.WriteString("没有水果")
	} else {
		firstFlag := true
		for variant, number := range fruitCounter {
			if firstFlag {
				firstFlag = false
				builder.WriteString("有")
			} else {
				builder.WriteString("、")
			}
			builder.WriteString(fmt.Sprintf("%d个%s", number, assets.GetFruitNameByVariant(variant)))
		}
	}
	builder.WriteString("，")
	if len(animalCounter) == 0 {
		builder.WriteString("没有动物")
	} else {
		firstFlag := true
		for _, variant := range animalCounter {
			if firstFlag {
				firstFlag = false
			} else {
				builder.WriteString("、")
			}
			builder.WriteString(fmt.Sprintf("一只%s", assets.GetAnimalNameByVariant(variant)))
		}
	}
	return builder.String()
}
