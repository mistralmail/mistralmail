package mistralmail

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/mistralmail/imap"
	"github.com/mistralmail/mistralmail/helpers"
	"github.com/mistralmail/smtp/server"
	log "github.com/sirupsen/logrus"
)

var (
	defaultSMTPPortIncoming      = 25
	defaultSMTPAddressIncoming   = fmt.Sprintf(":%d", defaultSMTPPortIncoming)
	defaultSMTPPortOutgoing      = 587
	defaultSMTPAddressOutgoing   = fmt.Sprintf(":%d", defaultSMTPPortOutgoing)
	defaultIMAPPort              = 143
	defaultIMAPAddress           = fmt.Sprintf(":%d", defaultIMAPPort)
	defaultHTTPPort              = 8080
	defaultHTTPAddress           = fmt.Sprintf(":%d", defaultHTTPPort)
	defaultDatabaseURL           = "sqlite:test.db"
	defaultAcmeEndpoint          = "https://acme-v02.api.letsencrypt.org/directory"
	defaultCertificatesDirectory = "./certificates"
)

// BuildConfigFromEnv populates a MistralMail config from env variables
func BuildConfigFromEnv() *Config {
	config := &Config{}

	// Core config
	config.Hostname = getEnv("HOSTNAME", "")
	config.SMTPAddressIncoming = getEnv("SMTP_ADDRESS_INCOMING", defaultSMTPAddressIncoming)
	config.SMTPAddressOutgoing = getEnv("SMTP_ADDRESS_OUTGOING", defaultSMTPAddressOutgoing)
	config.IMAPAddress = getEnv("IMAP_ADDRESS", defaultIMAPAddress)
	config.DatabaseURL = getEnv("DATABASE_URL", defaultDatabaseURL)

	outgoingMode := getEnv("SMTP_OUTGOING_MODE", "")
	if outgoingMode == string(SMTPOutgoingModeRelay) {
		config.SMTPOutgoingMode = SMTPOutgoingModeRelay
	}

	// SMTP external relay config
	config.ExternalRelayHostname = getEnv("EXTERNAL_RELAY_HOSTNAME", "")
	port := getEnv("EXTERNAL_RELAY_PORT", "")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		if port != "" {
			panic("TODO: return error here")
		}
	}
	config.ExternalRelayPort = portInt
	config.ExternalRelayUsername = getEnv("EXTERNAL_RELAY_USERNAME", "")
	config.ExternalRelayPassword = getEnv("EXTERNAL_RELAY_PASSWORD", "")

	skipVerify := getEnv("EXTERNAL_RELAY_INSECURE_SKIP_VERIFY", "")
	if strings.ToUpper(skipVerify) == "TRUE" {
		config.ExternalRelayInsecureSkipVerify = true
	}

	// TLS
	tlsDisable := getEnv("TLS_DISABLE", "")
	if strings.ToUpper(tlsDisable) == "TRUE" {
		config.DisableTLS = true
	}
	config.TLSCertificateFile = getEnv("TLS_CERTIFICATE_FILE", "")
	config.TLSPrivateKeyFile = getEnv("TLS_PRIVATE_KEY_FILE", "")
	config.AcmeEmail = getEnv("TLS_ACME_EMAIL", "")
	config.AcmeEndpoint = getEnv("TLS_ACME_ENDPOINT", defaultAcmeEndpoint)

	config.TLSCertificatesDirectory = getEnv("TLS_CERTIFICATES_DIRECTORY", defaultCertificatesDirectory)

	// HTTP
	config.HTTPAddress = getEnv("HTTP_ADDRESS", defaultHTTPAddress)

	config.Secret = getEnv("SECRET", "")

	return config
}

// SMTPOutgoingMode denotes the types of SMTP MSA modes.
type SMTPOutgoingMode string

const (
	// SMTPOutgoingModeRelay is the MSA Relay mode
	SMTPOutgoingModeRelay SMTPOutgoingMode = "RELAY"
)

// Config contains all the config for serving MistralMail
type Config struct {
	Hostname            string
	SMTPAddressIncoming string
	SMTPAddressOutgoing string
	SMTPOutgoingMode    SMTPOutgoingMode
	IMAPAddress         string
	HTTPAddress         string
	DatabaseURL         string
	Secret              string

	DisableTLS               bool
	TLSCertificatesDirectory string
	TLSCertificateFile       string
	TLSPrivateKeyFile        string
	AcmeEndpoint             string
	AcmeEmail                string

	ExternalRelayHostname           string
	ExternalRelayPort               int
	ExternalRelayUsername           string
	ExternalRelayPassword           string
	ExternalRelayInsecureSkipVerify bool
}

// Validate validates whether all config is set and valid
func (config *Config) Validate() error {

	// Core config
	if config.Hostname == "" {
		return fmt.Errorf("Hostname cannot be empty")
	}

	if config.SMTPAddressIncoming == "" {
		return fmt.Errorf("SMTPAddressIncoming cannot be empty")
	}

	if config.SMTPAddressOutgoing == "" {
		return fmt.Errorf("SMTPAddressOutgoing cannot be empty")
	}

	if config.SMTPOutgoingMode == "" {
		return fmt.Errorf("SMTPOutgoingMode cannot be empty")
	}
	if config.SMTPOutgoingMode != SMTPOutgoingModeRelay {
		return fmt.Errorf("unknown SMTPOutgoingMode")
	}

	if config.IMAPAddress == "" {
		return fmt.Errorf("IMAPAddress cannot be empty")
	}

	if config.DatabaseURL == "" {
		return fmt.Errorf("DatabaseURL cannot be empty")
	}

	// SMTP external relay config
	if config.SMTPOutgoingMode == SMTPOutgoingModeRelay {
		if config.ExternalRelayHostname == "" {
			return fmt.Errorf("ExternalRelayHostname cannot be empty when in relay mode")
		}
		if config.ExternalRelayPort == 0 {
			return fmt.Errorf("ExternalRelayPort cannot be empty when in relay mode")
		}
	}

	// TLS config
	if !config.DisableTLS {
		if config.TLSCertificateFile == "" && config.TLSPrivateKeyFile == "" {
			if config.AcmeEndpoint == "" {
				return fmt.Errorf("AcmeEndpoint should be defined when using TLS without providing certificates")
			}
			if config.AcmeEmail == "" {
				return fmt.Errorf("AcmeEmail should be defined when using TLS without providing certificates")
			}
			if config.TLSCertificatesDirectory == "" {
				return fmt.Errorf("TLSCertificatesDirectory should be defined when using TLS without providing certificates")
			}
		}
		if (config.TLSCertificateFile != "" && config.TLSPrivateKeyFile == "") || (config.TLSCertificateFile == "" && config.TLSPrivateKeyFile != "") {
			return fmt.Errorf("both TLSCertificateFile and TLSPrivateKeyFile must be defined when using custom TLS certificate")
		}
	}

	// HTTP server & api
	if config.HTTPAddress == "" {
		return fmt.Errorf("HTTPAddress cannot be empty")
	}
	if config.Secret == "" {
		return fmt.Errorf("Secret cannot be empty")
	}

	return nil
}

// GenerateMTAConfig generates the SMTP config for the MTA
func (config *Config) GenerateMTAConfig() *server.Config {

	host, port, err := net.SplitHostPort(config.SMTPAddressIncoming)
	if err != nil {
		// TODO: handle
		log.Fatalf("couldn't determine SMTP address/port: %v", err)
	}

	portInt, _ := strconv.Atoi(port)

	nixspamBlacklist, err := helpers.NewNixspam()
	if err != nil {
		log.Warnln("couldn't create Nixspam Blacklist instance: ", err)
	}

	return &server.Config{
		Hostname:    config.Hostname,
		Ip:          host,
		Port:        uint32(portInt),
		Blacklist:   nixspamBlacklist,
		DisableAuth: true,
	}
}

// GenerateMSAConfig generates the SMTP config for the MSA
func (config *Config) GenerateMSAConfig() *server.Config {

	host, port, err := net.SplitHostPort(config.SMTPAddressOutgoing)
	if err != nil {
		// TODO: handle
		log.Fatalf("couldn't determine SMTP address/port: %v", err)
	}

	portInt, _ := strconv.Atoi(port)

	nixspamBlacklist, err := helpers.NewNixspam()
	if err != nil {
		log.Warnln("couldn't create Nixspam Blacklist instance: ", err)
	}

	return &server.Config{
		Hostname:    config.Hostname,
		Ip:          host,
		Port:        uint32(portInt),
		Blacklist:   nixspamBlacklist,
		DisableAuth: false,
	}
}

// GenerateIMAPBackendConfig generates the config object for the IMAP backend
func (config *Config) GenerateIMAPBackendConfig() *imap.Config {
	return &imap.Config{
		IMAPAddress: config.IMAPAddress,
	}
}

// getEnv gets the env variable with the given key if the key exists
// else it falls back to the fallback value
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
