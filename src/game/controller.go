package game

import (
	"halligalli/model"
	"log"
	"time"
)

type EventType = int

type Event struct {
	EventType EventType
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

func InitiateGame(game *Game, messageChannel chan Message) {
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
	game := &Game{}
	game.Init()
	ticker := time.NewTicker(game.Rule.DealInterval)
	ticker.Stop()
	for {
		select {
		case event := <-eventChannel:
			switch event.EventType {
			case Initiate:
				if game.State == Closed || game.State == WaitingForStart {
					InitiateGame(game, messageChannel)
				}
			case Start:
				if game.State == WaitingForStart {
					game.State = Running
					ticker.Reset(game.Rule.DealInterval)

					RevealCardAndSend(game, messageChannel)
				}
			case RingTheBell:
				if game.State == Running {
					ticker.Stop()
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
					ticker.Reset(game.Rule.DealInterval)
					game.State = Running
					RevealCardAndSend(game, messageChannel)
				}
			case Terminate:
				if game.State == WaitingForStart || game.State == Running || game.State == Paused {
					ticker.Stop()
					TerminateGame(game, messageChannel)
				}
			}
		case <-ticker.C:
			if game.State == Running {
				RevealCardAndSend(game, messageChannel)
			}
		}
	}
}
