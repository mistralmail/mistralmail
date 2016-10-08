package handlers

import (
	"github.com/gopistolet/gopistolet/handlers/maildir"
	"github.com/gopistolet/gopistolet/handlers/spf"
)

// LoadHandlers creates a HandlerMechanism object with the needed/available loaders
func LoadHandlers() *HandlerMachanism {
	return &HandlerMachanism{
		Handlers: []Handler{
    		spf.New(),
			maildir.New(),
		},
	}
}
