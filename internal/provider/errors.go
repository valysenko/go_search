package provider

import "fmt"

type Provider string

const (
	DevTo    Provider = "devto"
	Wiki     Provider = "wiki"
	Hashnode Provider = "hashnode"
)

type ProviderError struct {
	Msg      string
	Provider Provider
	Err      error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Provider, e.Msg, e.Err)
}
