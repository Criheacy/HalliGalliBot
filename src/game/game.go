package game

import (
	"halligalli/assets"
	"halligalli/common"
	"halligalli/env"
	"log"
	"math/rand"
	"time"
)

type State = int

const (
	Closed State = iota
	WaitingForStart
	Paused
	Running
)

type Game struct {
	ChannelId     string
	Round         int
	State         State
	Deck          []common.Card
	NextCardIndex int
	RevealTimer   *time.Timer
	RevealedCards []common.Card
}

func (game *Game) Init(channelId string) {
	game.ChannelId = channelId
	game.State = Closed
	game.Deck = make([]common.Card, len(env.GetContext().Asset.Cards))
	copy(game.Deck, env.GetContext().Asset.Cards)
	game.ShuffleDeck()
	game.NextCardIndex = 0
	game.RevealTimer = time.NewTimer(env.GetContext().GameRule.DealInterval)
	game.RevealTimer.Stop()
	game.RevealedCards = make([]common.Card, 0)
}

func (game *Game) ShuffleDeck() {
	for i := range game.Deck {
		j := rand.Intn(i + 1)
		game.Deck[i], game.Deck[j] = game.Deck[j], game.Deck[i]
	}
}

func (game *Game) RevealNextCard() common.Card {
	if game.NextCardIndex >= len(game.Deck) {
		game.NextCardIndex = 0
		game.ShuffleDeck()
	}
	card := game.Deck[game.NextCardIndex]
	if game.RevealedCards == nil {
		game.RevealedCards = make([]common.Card, 0)
	}
	game.RevealedCards = append(game.RevealedCards, card)
	game.NextCardIndex += 1
	return card
}

type Counter struct {
	Variant int
	Count   int
}

func GetFruitCounters() []Counter {
	result := make([]Counter, len(env.GetContext().Asset.Meta.Fruits))
	for index, fruit := range env.GetContext().Asset.Meta.Fruits {
		result[index].Variant = fruit.Variant
		result[index].Count = 0
	}
	return result
}

func CountFruit(variant int, number int, fruitCounters *[]Counter) {
	for _, fruitCounter := range *fruitCounters {
		if fruitCounter.Variant == variant {
			fruitCounter.Count += number
		}
	}
}

// WinCheck returns (isWin, animalName, fruitName)
func (game *Game) WinCheck() (bool, string, string) {
	validCards := game.GetValidCards()

	hasAnimal := false
	animalVariant := -1
	fruitCounters := GetFruitCounters()

	for _, card := range validCards {
		if card.Type == common.Fruit {
			for _, element := range card.Elements {
				CountFruit(element.Variant, element.Number, &fruitCounters)
			}
		} else if card.Type == common.Animal {
			hasAnimal = true
			animalVariant = card.Variant
		}
	}
	log.Printf("has animal: %t", hasAnimal)

	if hasAnimal {
		animalName := assets.GetAnimalNameByVariant(animalVariant)
		return true, animalName, ""
	}

	for _, counter := range fruitCounters {
		log.Printf("fruit counter: %d %d", counter.Variant, counter.Count)
		if counter.Count == env.GetContext().GameRule.FruitNumberToWin {
			fruitName := assets.GetFruitNameByVariant(counter.Variant)
			return true, "", fruitName
		}
	}

	return false, "", ""
}

func (game *Game) GetValidCards() []common.Card {
	sliceFrom := maxInt(0, len(game.RevealedCards)-env.GetContext().GameRule.ValidCardNumber)
	validCards := game.RevealedCards[sliceFrom:]
	return validCards
}

func (game *Game) NewRound() {
	game.RevealedCards = nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
