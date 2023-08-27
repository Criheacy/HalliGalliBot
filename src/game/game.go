package game

import (
	"encoding/json"
	"fmt"
	"halligalli/assets"
	"halligalli/auth"
	"halligalli/env"
	"log"
	"math/rand"
	"os"
	"time"
)

const AssetFilePath = "./asset.json"

type Rule struct {
	ValidCardNumber  int
	FruitNumberToWin int
	DealInterval     time.Duration
}

type Card struct {
	Image    string        `json:"image"`
	Type     CardType      `json:"type"`
	Variant  int           `json:"variant"`
	Repeat   int           `json:"repeat"`
	Elements []CardElement `json:"elements"`
}

type CardType = string

const (
	Fruit  CardType = "fruit"
	Animal CardType = "animal"
)

type CardElement struct {
	Variant int `json:"variant"`
	Number  int `json:"number"`
}

type Asset struct {
	Meta  AssetMeta `json:"meta"`
	Cards []Card    `json:"cards"`
}

type AssetVariant struct {
	Name    string `json:"name"`
	Variant int    `json:"variant"`
}

type AssetMeta struct {
	Fruits  []AssetVariant `json:"fruits"`
	Animals []AssetVariant `json:"animals"`
}

type State = int

const (
	Closed State = iota
	WaitingForStart
	Paused
	Running
)

type Game struct {
	Rule          Rule
	Round         int
	State         State
	Deck          []Card
	NextCard      int
	RevealedCards []Card
}

func LoadAssets() error {
	content, err := os.ReadFile(auth.GetPath(AssetFilePath))
	if err != nil {
		return err
	}
	var asset Asset
	if err = json.Unmarshal(content, &asset); err != nil {
		return err
	}
	env.GetContext().Asset = asset
	return nil
}

func (game *Game) Init() {
	game.Rule = Rule{
		ValidCardNumber:  5,
		FruitNumberToWin: 5,
		DealInterval:     7 * time.Second,
	}
	game.State = Closed
	game.Deck = make([]Card, len(env.GetContext().Asset.Cards))
	copy(game.Deck, env.GetContext().Asset.Cards)
	game.ShuffleDeck()
	game.NextCard = 0
	game.RevealedCards = make([]Card, 0)
}

func (game *Game) ShuffleDeck() {
	for i := range game.Deck {
		j := rand.Intn(i + 1)
		game.Deck[i], game.Deck[j] = game.Deck[j], game.Deck[i]
	}
}

func (game *Game) RevealNextCard() Card {
	if game.NextCard >= len(game.Deck) {
		game.NextCard = 0
		game.ShuffleDeck()
	}
	card := game.Deck[game.NextCard]
	if game.RevealedCards == nil {
		game.RevealedCards = make([]Card, 0)
	}
	game.RevealedCards = append(game.RevealedCards, card)
	game.NextCard += 1
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
	sliceFrom := maxInt(0, len(game.RevealedCards)-game.Rule.ValidCardNumber)
	validCards := game.RevealedCards[sliceFrom:]

	var cardLog string
	for _, card := range validCards {
		if card.Type == Animal {
			cardLog += fmt.Sprintf("[%s] ", assets.GetAnimalNameByVariant(card.Variant))
		} else {
			for _, element := range card.Elements {
				cardLog += fmt.Sprintf("[%s x%d] ", assets.GetFruitNameByVariant(element.Variant), element.Number)
			}
		}
	}
	log.Printf("valid cards: %s", cardLog)

	hasAnimal := false
	animalVariant := -1
	fruitCounters := GetFruitCounters()

	for _, card := range validCards {
		if card.Type == Fruit {
			for _, element := range card.Elements {
				CountFruit(element.Variant, element.Number, &fruitCounters)
			}
		} else if card.Type == Animal {
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
		if counter.Count == game.Rule.FruitNumberToWin {
			fruitName := assets.GetFruitNameByVariant(counter.Variant)
			return true, "", fruitName
		}
	}

	return false, "", ""
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
