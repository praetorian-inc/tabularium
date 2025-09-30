package model

var configurationValidators = map[string]func(value any) error{
	"subscription": subscriptionValidator,
}
