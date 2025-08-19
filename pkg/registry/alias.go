package registry

type Alias interface {
	SetAlias(string)
}

type ModelAlias struct {
	Alias string
}

func (m *ModelAlias) SetAlias(alias string) {
	m.Alias = alias
}
