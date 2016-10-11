package spf

import (
	"fmt"
	"strings"

	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gospf"
	"github.com/gopistolet/gospf/dns"
	"github.com/gopistolet/smtp/mta"
	"github.com/gopistolet/smtp/smtp"
)

func New(c *mta.Config) *Spf {
	return &Spf{
		config: c,
	}
}

type Spf struct {
	config *mta.Config
}

func (handler *Spf) Handle(state *smtp.State) {
	// create SPF instance
	spf, err := gospf.New(state.From.GetDomain(), &dns.GoSPFDNS{})
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Infof("Could not create spf: %v", err)
		return
	}

	// check the given IP on that instance
	check, err := spf.CheckIP(state.Ip.String())
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Errorf("Error while checking ip in spf: %v", err)
		return
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

}
