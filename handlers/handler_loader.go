package handlers

import (
	"github.com/gopistolet/gopistolet/handlers/received"
	"github.com/gopistolet/gopistolet/handlers/spf"
	"github.com/gopistolet/smtp/mta"
)

// LoadHandlers creates a HandlerMechanism object with the needed/available loaders
func LoadHandlers(c *mta.Config) *HandlerMachanism {
	return &HandlerMachanism{
		Handlers: []Handler{
			received.New(c),
			spf.New(c),
		},
	}
}
