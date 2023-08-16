package handlers

import (
	"github.com/mistralmail/mistralmail/handlers/received"
	"github.com/mistralmail/mistralmail/handlers/spf"
	"github.com/mistralmail/smtp/server"
)

// LoadHandlers creates a HandlerMechanism object with the needed/available loaders
func LoadHandlers(c *server.Config) *HandlerMachanism {
	return &HandlerMachanism{
		Handlers: []Handler{
			received.New(c),
			spf.New(c),
		},
	}
}
