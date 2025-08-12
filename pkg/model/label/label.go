package label

func New(value string) string {
	registry := GetRegistry()
	registry.mustRegister(value)
	return value
}
