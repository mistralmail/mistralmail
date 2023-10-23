package received

import (
	"fmt"
	"time"

	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"
	log "github.com/sirupsen/logrus"
)

func New(c *server.Config) *Received {
	return &Received{
		config: c,
	}
}

type Received struct {
	config *server.Config
}

func (handler *Received) Handle(state *smtp.State) error {

	/*
	   RFC 2076 3.2 Trace information

	       Trace of MTAs which a message has passed.


	   RFC 5322 3.6.7.

	       received        =   "Received:" *received-token ";" date-time CRLF

	       received-token  =   word / angle-addr / addr-spec / domain


	   Example:

	       Received: from mail.example.com (192.168.0.10) by some.mail.server.example.com (192.168.0.11) with Microsoft SMTP Server id 14.3.319.2; Wed, 5 Oct 2016 14:57:46 +0200
	*/
	headerKey := "Received"

	date := time.Now().Format(time.RFC1123Z) // date-time in RFC 5322 is like RFC 1123Z
	headerValue := fmt.Sprintf("from %s (%s) by %s (%s) with MistralMail; %s", state.Hostname, state.Ip, handler.config.Hostname, handler.config.Ip, date)

	state.AddHeader(headerKey, headerValue)

	// TODO: 'by IP' is not necessarily set in config

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
		"Hostname":  state.Hostname,
	}).Debug("Added 'received' header: '", headerValue, "'")

	return nil
}
