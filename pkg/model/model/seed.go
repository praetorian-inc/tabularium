package model

const SeedLabel = "Seed"

type Seedable interface {
	Target
	SeedModels() []Seedable
	GetSource() string
	SetSource(string)
	SetOrigin(string)
}
