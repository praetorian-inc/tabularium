package model

type CredentialFormat string

const (
	FormatEnv   CredentialFormat = "env"
	FormatFile  CredentialFormat = "file"
	FormatToken CredentialFormat = "token"
)

type CredentialFormatter interface {
	Apply(CredentialResponse, Job) error
	Cleanup() error
}
