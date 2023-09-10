package relay

import (
	"crypto/tls"
	"fmt"

	"github.com/mistralmail/smtp/smtp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"

	netSmtp "net/smtp"
)

// Define a counter vector for successful SMTP deliveries.
var smtpDelivered = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "smtp_delivered",
		Help: "SMTP deliveries status (success or error)",
	},
	[]string{"status"}, // The label "status" can have values "success" or "error".
)

// New creates a new Relay handler.
func New(relayHostname string, relayPort int, relayUsername string, relayPassword string, relayInsecureSkipVerify bool) *Relay {
	return &Relay{
		relayHostname:           relayHostname,
		relayPort:               relayPort,
		relayUsername:           relayUsername,
		relayPassword:           relayPassword,
		relayInsecureSkipVerify: relayInsecureSkipVerify,
	}
}

// Relay is an SMTP handler that forwards the incoming mail to another SMTP relay server.
type Relay struct {
	relayHostname           string
	relayPort               int
	relayUsername           string
	relayPassword           string
	relayInsecureSkipVerify bool
}

// Handle handles the state.
func (handler *Relay) Handle(state *smtp.State) error {

	recipients := make([]string, len(state.To))

	for i, to := range state.To {
		if to == nil {
			// maybe this shouldn't even be a pointer at all?
			return fmt.Errorf("state.To cannot be nil")
		}
		recipients[i] = to.Address
	}

	err := handler.SendMail(state.From.Address, recipients, state.Data)
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
			"Hostname":  state.Hostname,
		}).Errorf("Couldn't deliver message to relay: %v", err)

		// Increment the "error" counter.
		smtpDelivered.WithLabelValues("error").Inc()

		return err
	}

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
		"Hostname":  state.Hostname,
	}).Debug("Delived message to relay")

	// Increment the "success" counter.
	smtpDelivered.WithLabelValues("success").Inc()

	return nil
}

// SendMail sends an SMTP message with a given from and to mail address.
// code partially copy pasted from the net/smtp package to allow setting a custom TLS config.
// should be moved to the SMTP package
func (handler *Relay) SendMail(from string, recipients []string, message []byte) error {

	// Connect to the SMTP server without TLS
	smtpAddress := fmt.Sprintf("%s:%d", handler.relayHostname, handler.relayPort)
	conn, err := netSmtp.Dial(smtpAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to the SMTP server: %w", err)
	}
	defer conn.Close()

	// Create TLS config
	config := &tls.Config{
		ServerName:         handler.relayHostname,
		InsecureSkipVerify: handler.relayInsecureSkipVerify, // disable TLS if configured that way
	}
	err = conn.StartTLS(config)
	if err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if handler.relayUsername != "" {
		auth := netSmtp.PlainAuth("", handler.relayUsername, handler.relayPassword, handler.relayHostname)
		err = conn.Auth(auth)
		if err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set the sender and recipient
	err = conn.Mail(from)
	if err != nil {
		return fmt.Errorf("failed to set the sender: %w", err)
	}

	for _, to := range recipients {
		err = conn.Rcpt(to)
		if err != nil {
			return fmt.Errorf("failed to set the recipient: %w", err)
		}
	}

	// Send the email
	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to send the email: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close the data connection: %v", err)
	}

	return nil
}
