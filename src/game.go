package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

const AssetFilePath = "./asset.json"

type GameRule struct {
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
	Variant string `json:"variant"`
}

type AssetMeta struct {
	Fruits  []AssetVariant `json:"fruits"`
	Animals []AssetVariant `json:"animals"`
}

type GameState = int

const (
	Closed GameState = iota
	WaitingForStart
	Paused
	Running
)

type Game struct {
	Rule          GameRule
	Round         int
	State         GameState
	Deck          []Card
	NextCard      int
	RevealedCards []Card
}

func LoadAssets() error {
	content, err := os.ReadFile(GetPath(AssetFilePath))
	if err != nil {
		return err
	}
	var asset Asset
	if err = json.Unmarshal(content, &asset); err != nil {
		return err
	}
	GetContext().Asset = asset
	return nil
}

func (game *Game) Init() {
	game.Rule = GameRule{
		ValidCardNumber:  5,
		FruitNumberToWin: 5,
		DealInterval:     10 * time.Second,
	}
	game.State = Closed
	game.Deck = make([]Card, len(GetContext().Asset.Cards))
	copy(game.Deck, GetContext().Asset.Cards)
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

// WinCheck returns (isWin, animalVariant(defaults -1), fruitVariant(defaults -1))
func (game *Game) WinCheck() (bool, int, int) {
	sliceFrom := maxInt(0, len(game.RevealedCards)-game.Rule.ValidCardNumber)
	validCards := game.RevealedCards[sliceFrom:]

	log.Printf("valid cards: %+v", validCards)

	hasAnimal := false
	animalVariant := -1
	fruitCount := make([]int, len(GetContext().Asset.Meta.Fruits))
	for _, card := range validCards {
		if card.Type == Fruit {
			for _, element := range card.Elements {
				fruitCount[element.Variant] += element.Number
			}
		} else if card.Type == Animal {
			hasAnimal = true
			animalVariant = card.Variant
		}
	}

	if hasAnimal {
		return true, animalVariant, -1
	}

	for variant, count := range fruitCount {
		if count == game.Rule.FruitNumberToWin {
			return true, -1, variant
		}
	}

	return false, -1, -1
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
