package handlers

import (
	"github.com/mistralmail/smtp/smtp"
)

// Handler is an interface for SMTP handlers.
type Handler interface {
	// Handle is a handler that will be called by the SMTP server with the given SMTP state.
	Handle(state *smtp.State) error
}

// HandlerMechanism contains a list of all handlers and executes the chain
// it is meant to be passed to the MTA as mta.Handler interface
type HandlerMachanism struct {
	Handlers []Handler
}

// AddHandler adds a handler to the mechanism.
func (h *HandlerMachanism) AddHandler(handler ...Handler) {
	h.Handlers = append(h.Handlers, handler...)
}

// Handle implements the Handler interface.
func (h *HandlerMachanism) Handle(state *smtp.State) error {
	for _, handler := range h.Handlers {
		err := handler.Handle(state)
		if err != nil {
			return err
		}
	}
	return nil
}
