package main

func GetAnimalNameByVariant(variant int) string {
	for _, animal := range GetContext().Asset.Meta.Animals {
		if animal.Variant == variant {
			return animal.Name
		}
	}
	return ""
}

func GetFruitNameByVariant(variant int) string {
	for _, fruit := range GetContext().Asset.Meta.Fruits {
		if fruit.Variant == variant {
			return fruit.Name
		}
	}
	return ""
}
