package main

import "time"

type GameEvent = int

const (
	Initiate GameEvent = iota
	Start
	RingTheBell
	Continue
	Terminate
)

type GameMessageType = int

type GameMessage struct {
	MessageType GameMessageType
	Param       any
}

const (
	ShowGameRule GameMessageType = iota
	CardRevealed
	PlayerWin
	FakeRing
	GameTerminated
)

type RoundStatus struct {
	Win           bool
	AnimalVariant int
	FruitVariant  int
}

func InitiateGame(game Game, messageChannel chan GameMessage) {
	messageChannel <- GameMessage{
		MessageType: ShowGameRule,
		Param:       nil,
	}
	game.State = WaitingForStart
}

func RevealCardAndSend(game Game, messageChannel chan GameMessage) {
	card := game.RevealNextCard()
	messageChannel <- GameMessage{
		MessageType: CardRevealed,
		Param:       card,
	}
}

func TerminateGame(game Game, messageChannel chan GameMessage) {
	card := game.RevealNextCard()
	messageChannel <- GameMessage{
		MessageType: CardRevealed,
		Param:       card,
	}
}

func GameLoop(eventChannel chan GameEvent, messageChannel chan GameMessage) {
	game := Game{}
	ticker := time.NewTicker(game.Rule.DealInterval)
	ticker.Stop()
	for {
		select {
		case event := <-eventChannel:
			switch event {
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
					isWin, animalVariant, fruitVariant := game.WinCheck()
					roundStatus := RoundStatus{
						Win:           isWin,
						AnimalVariant: animalVariant,
						FruitVariant:  fruitVariant,
					}
					if isWin {
						messageChannel <- GameMessage{
							MessageType: PlayerWin,
							Param:       roundStatus,
						}
						game.NewRound()
					} else {
						messageChannel <- GameMessage{
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
