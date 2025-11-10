package model

const SeedLabel = "Seed"

type Seedable interface {
	Target
	SeedModels() []Seedable
	GetSource() string
	GetOrigin() string
	SetSource(string)
	SetOrigin(string)
}
