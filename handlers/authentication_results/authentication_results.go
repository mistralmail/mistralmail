package authenticationresults

import (
	"fmt"
	"strings"

	"github.com/mistralmail/gospf"
	"github.com/mistralmail/gospf/dns"
	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"
	log "github.com/sirupsen/logrus"
)

// New creates a new AuthenticationResults handler.
// currently it only checks for SPF.
func New(c *server.Config) *AuthenticationResults {
	return &AuthenticationResults{
		config: c,
	}
}

// AuthenticationResults handler adds the 'Authentication-Results' header.
type AuthenticationResults struct {
	config *server.Config
}

// Handle implments the HandlerFunc interface.
func (handler *AuthenticationResults) Handle(state *smtp.State) error {

	// create SPF instance
	spf, err := gospf.New(state.From.GetDomain(), &dns.GoSPFDNS{})
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Infof("Could not create spf: %v", err)
		return nil
	}

	// check the given IP on that instance
	check, err := spf.CheckIP(state.Ip.String())
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Errorf("Error while checking ip in spf: %v", err)
		return nil
	}
	log.WithFields(log.Fields{
		"Ip":     state.Ip.String(),
		"Domain": state.From.GetDomain(),
	}).Info("SPF returned " + check)

	// write Authentication-Results header
	// TODO: need value from config here...
	//
	// header field is defined in RFC 5451 section 2.2
	// Authentication-Results: receiver.example.org; spf=pass smtp.mailfrom=example.com;
	headerField := fmt.Sprintf("Authentication-Results: %s; spf=%s smtp.mailfrom=%s;\r\n", handler.config.Hostname, strings.ToLower(check), state.From.GetDomain())
	state.Data = append([]byte(headerField), state.Data...)

	return nil

}
