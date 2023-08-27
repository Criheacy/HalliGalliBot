package common

import (
	"time"
)

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

type Token struct {
	AppID       uint64
	AccessToken string
}

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
