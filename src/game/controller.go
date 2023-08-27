package game

import (
	"halligalli/env"
	"halligalli/model"
	"log"
)

type EventType = int

type Event struct {
	EventType EventType
	ChannelId string
	Param     any
}

const (
	Initiate EventType = iota
	Start
	RingTheBell
	Continue
	Terminate
)

type MessageType = int

type Message struct {
	MessageType MessageType
	ChannelId   string
	Param       any
}

const (
	ShowGameRule MessageType = iota
	CardRevealed
	PlayerWin
	FakeRing
	Terminated
)

type RoundStatus struct {
	IsWin      bool
	Player     model.User
	AnimalName string
	FruitName  string
}

type RevealTickerEvent struct {
	Game *Game
}

func Initiated(game *Game, messageChannel chan Message) {
	messageChannel <- Message{
		MessageType: ShowGameRule,
		Param:       nil,
	}
	game.State = WaitingForStart
}

func RevealCardAndSend(game *Game, messageChannel chan Message) {
	card := game.RevealNextCard()
	log.Printf("card revealed: %+v", card)
	messageChannel <- Message{
		MessageType: CardRevealed,
		Param:       card,
	}
}

func TerminateGame(game *Game, messageChannel chan Message) {
	game.State = Closed
	messageChannel <- Message{
		MessageType: Terminated,
		Param:       nil,
	}
}

func MainLoop(eventChannel chan Event, messageChannel chan Message) {
	gameInstances := make(map[string]*Game)
	tickerChannel := make(chan RevealTickerEvent, 32)
	for {
		select {
		case event := <-eventChannel:
			game := gameInstances[event.ChannelId]
			if game == nil {
				game = &Game{}
				game.Init(event.ChannelId)
				gameInstances[event.ChannelId] = game
			}
			switch event.EventType {
			case Initiate:
				if game.State == Closed || game.State == WaitingForStart {
					Initiated(game, messageChannel)
				}
			case Start:
				if game.State == WaitingForStart {
					game.State = Running
					RevealCardAndSend(game, messageChannel)
					go WaitForNextCard(game, tickerChannel)
				}
			case RingTheBell:
				if game.State == Running {
					game.RevealTimer.Stop()
					game.State = Paused
					isWin, animalName, fruitName := game.WinCheck()
					roundStatus := RoundStatus{
						IsWin:      isWin,
						Player:     event.Param.(model.User),
						AnimalName: animalName,
						FruitName:  fruitName,
					}
					if isWin {
						messageChannel <- Message{
							MessageType: PlayerWin,
							Param:       roundStatus,
						}
						game.NewRound()
					} else {
						messageChannel <- Message{
							MessageType: FakeRing,
							Param:       roundStatus,
						}
					}
				}
			case Continue:
				if game.State == Paused {
					game.State = Running
					RevealCardAndSend(game, messageChannel)
					go WaitForNextCard(game, tickerChannel)
				}
			case Terminate:
				if game.State == WaitingForStart || game.State == Running || game.State == Paused {
					game.RevealTimer.Stop()
					TerminateGame(game, messageChannel)
				}
			}
		case tickerEvent := <-tickerChannel:
			game := tickerEvent.Game
			RevealCardAndSend(game, messageChannel)
			go WaitForNextCard(game, tickerChannel)
		}
	}
}

func WaitForNextCard(game *Game, tickerChannel chan RevealTickerEvent) {
	game.RevealTimer.Reset(env.GetContext().GameRule.DealInterval)
	for {
		select {
		case <-game.RevealTimer.C:
			if game.State == Running {
				tickerChannel <- RevealTickerEvent{
					Game: game,
				}
			}
			break
		}
	}
}
