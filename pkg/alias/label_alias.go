package alias

type LabelAliaser interface {
	SetLabelAlias(string)
}

type LabelAlias struct {
	Alias string `neo4j:"-" json:"-"`
}

func (m *LabelAlias) SetLabelAlias(alias string) {
	m.Alias = alias
}
