package registry

type Alias interface {
	SetAlias(string)
}

type ModelAlias struct {
	Alias string `neo4j:"-" json:"-"`
}

func (m *ModelAlias) SetAlias(alias string) {
	m.Alias = alias
}
