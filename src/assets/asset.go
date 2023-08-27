package assets

import "halligalli/env"

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
