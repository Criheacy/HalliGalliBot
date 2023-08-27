package assets

import (
	"encoding/json"
	"halligalli/auth"
	"halligalli/common"
	"halligalli/env"
	"os"
)

const AssetFilePath = "./asset.json"

func LoadAssets() error {
	content, err := os.ReadFile(auth.GetPath(AssetFilePath))
	if err != nil {
		return err
	}
	var asset common.Asset
	if err = json.Unmarshal(content, &asset); err != nil {
		return err
	}
	env.GetContext().Asset = asset
	return nil
}

func GetAnimalNameByVariant(variant int) string {
	for _, animal := range env.GetContext().Asset.Meta.Animals {
		if animal.Variant == variant {
			return animal.Name
		}
	}
	return ""
}

func GetFruitNameByVariant(variant int) string {
	for _, fruit := range env.GetContext().Asset.Meta.Fruits {
		if fruit.Variant == variant {
			return fruit.Name
		}
	}
	return ""
}
