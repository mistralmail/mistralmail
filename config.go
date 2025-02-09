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
	defaultSubDomainIncoming     = "mx"
	defaultSMTPPortOutgoing      = 587
	defaultSMTPAddressOutgoing   = fmt.Sprintf(":%d", defaultSMTPPortOutgoing)
	defaultSubDomainOutgoing     = "smtp"
	defaultIMAPPort              = 143
	defaultIMAPAddress           = fmt.Sprintf(":%d", defaultIMAPPort)
	defaultIMAPSubdomain         = "imap"
	defaultHTTPPort              = 8080
	defaultHTTPAddress           = fmt.Sprintf(":%d", defaultHTTPPort)
	defaultMetricsPort           = 9000
	defaultMetricsAddress        = fmt.Sprintf(":%d", defaultMetricsPort)
	defaultDatabaseURL           = "sqlite:test.db"
	defaultAcmeEndpoint          = "https://acme-v02.api.letsencrypt.org/directory"
	defaultCertificatesDirectory = "./certificates"
	defaultBlacklistURL          = "https://raw.githubusercontent.com/bitwire-it/ipblocklist/refs/heads/main/ip-list.txt"
)

// BuildConfigFromEnv populates a MistralMail config from env variables
func BuildConfigFromEnv() (*Config, error) {
	config := &Config{}

	// Core config
	config.Hostname = getEnv("HOSTNAME", "")
	config.SMTPAddressIncoming = getEnv("SMTP_ADDRESS_INCOMING", defaultSMTPAddressIncoming)
	config.SMTPAddressOutgoing = getEnv("SMTP_ADDRESS_OUTGOING", defaultSMTPAddressOutgoing)
	config.IMAPAddress = getEnv("IMAP_ADDRESS", defaultIMAPAddress)
	config.DatabaseURL = getEnv("DATABASE_URL", defaultDatabaseURL)

	outgoingMode := getEnv("SMTP_OUTGOING_MODE", "")
	if strings.ToUpper(outgoingMode) == string(SMTPOutgoingModeRelay) {
		config.SMTPOutgoingMode = SMTPOutgoingModeRelay
	}

	config.SubDomainIncoming = getEnv("SUBDOMAIN_INCOMING", fmt.Sprintf("%s.%s", defaultSubDomainIncoming, config.Hostname))
	config.SubDomainOutgoing = getEnv("SUBDOMAIN_OUTGOING", fmt.Sprintf("%s.%s", defaultSubDomainOutgoing, config.Hostname))
	config.SubDomainIMAP = getEnv("SUBDOMAIN_IMAP", fmt.Sprintf("%s.%s", defaultIMAPSubdomain, config.Hostname))

	// SMTP external relay config
	config.ExternalRelayHostname = getEnv("EXTERNAL_RELAY_HOSTNAME", "")
	port := getEnv("EXTERNAL_RELAY_PORT", "")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		if port != "" {
			return nil, fmt.Errorf("couldn't parse EXTERNAL_RELAY_PORT")
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
	acmeChallenge := getEnv("TLS_ACME_CHALLENGE", "")
	if strings.ToUpper(acmeChallenge) == string(AcmeChallengeHTTP) {
		config.AcmeChallenge = AcmeChallengeHTTP
	}
	if strings.ToUpper(acmeChallenge) == string(AcmeChallengeDNS) {
		config.AcmeChallenge = AcmeChallengeDNS
	}
	config.AcmeEmail = getEnv("TLS_ACME_EMAIL", "")
	config.AcmeEndpoint = getEnv("TLS_ACME_ENDPOINT", defaultAcmeEndpoint)
	config.AcmeDNSProvider = getEnv("TLS_ACME_DNS_PROVIDER", "")

	config.TLSCertificatesDirectory = getEnv("TLS_CERTIFICATES_DIRECTORY", defaultCertificatesDirectory)

	// HTTP
	config.HTTPAddress = getEnv("HTTP_ADDRESS", defaultHTTPAddress)

	config.Secret = getEnv("SECRET", "")

	// Metrics
	config.MetricsAddress = getEnv("METRICS_ADDRESS", defaultMetricsAddress)

	// Sentry
	config.SentryDSN = getEnv("SENTRY_DSN", "")

	// Log full queries
	logFullQueries := getEnv("LOG_FULL_QUERIES", "")
	if strings.ToUpper(logFullQueries) == "TRUE" {
		config.LogFullQueries = true
	}

	// Spam check
	spamCheckEnable := getEnv("SPAM_CHECK_ENABLE", "")
	if strings.ToUpper(spamCheckEnable) == "TRUE" {
		config.EnableSpamCheck = true
	}

	// Blacklist URL
	config.BlacklistURL = getEnv("BLACKLIST_URL", defaultBlacklistURL)

	return config, nil
}

// SMTPOutgoingMode denotes the types of SMTP MSA modes.
type SMTPOutgoingMode string

const (
	// SMTPOutgoingModeRelay is the MSA Relay mode
	SMTPOutgoingModeRelay SMTPOutgoingMode = "RELAY"
)

// AcmeChallenge denotes the types of Let's Encrypt challenges
type AcmeChallenge string

const (
	// AcmeChallengeHTTP is the standard HTTP-01 or TLS-ALPN-01 challenge.
	AcmeChallengeHTTP AcmeChallenge = "HTTP"
	// AcmeChallengeDNS is the DNS-01 challenge.
	AcmeChallengeDNS AcmeChallenge = "DNS"
)

// Config contains all the config for serving MistralMail
type Config struct {
	Hostname            string
	SubDomainIncoming   string
	SMTPAddressIncoming string
	SubDomainOutgoing   string
	SMTPAddressOutgoing string
	SMTPOutgoingMode    SMTPOutgoingMode
	SubDomainIMAP       string
	IMAPAddress         string
	HTTPAddress         string
	DatabaseURL         string
	Secret              string
	MetricsAddress      string
	SentryDSN           string
	LogFullQueries      bool
	EnableSpamCheck     bool
	BlacklistURL        string

	DisableTLS               bool
	TLSCertificatesDirectory string
	TLSCertificateFile       string
	TLSPrivateKeyFile        string
	AcmeChallenge            AcmeChallenge
	AcmeEndpoint             string
	AcmeEmail                string
	AcmeDNSProvider          string

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
		return fmt.Errorf("HOSTNAME cannot be empty")
	}

	if config.SMTPAddressIncoming == "" {
		return fmt.Errorf("SMTP_ADDRESS_INCOMING cannot be empty")
	}

	if config.SMTPAddressOutgoing == "" {
		return fmt.Errorf("SMTP_ADDRESS_OUTGOING cannot be empty")
	}

	if config.SMTPOutgoingMode == "" {
		return fmt.Errorf("SMTP_OUTGOING_MODE cannot be empty")
	}
	if config.SMTPOutgoingMode != SMTPOutgoingModeRelay {
		return fmt.Errorf("unknown SMTP_OUTGOING_MODE")
	}

	if config.IMAPAddress == "" {
		return fmt.Errorf("IMAP_ADDRESS cannot be empty")
	}

	if config.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL cannot be empty")
	}

	if config.SubDomainIncoming == "" {
		return fmt.Errorf("SUBDOMAIN_INCOMING cannot be empty")
	}

	if config.SubDomainOutgoing == "" {
		return fmt.Errorf("SUBDOMAIN_OUTGOING cannot be empty")
	}

	if config.SubDomainIMAP == "" {
		return fmt.Errorf("SUBDOMAIN_IMAP cannot be empty")
	}

	// SMTP external relay config
	if config.SMTPOutgoingMode == SMTPOutgoingModeRelay {
		if config.ExternalRelayHostname == "" {
			return fmt.Errorf("EXTERNAL_RELAY_HOSTNAME cannot be empty when in relay mode")
		}
		if config.ExternalRelayPort == 0 {
			return fmt.Errorf("EXTERNAL_RELAY_PORT cannot be empty when in relay mode")
		}
	}

	// TLS config
	if !config.DisableTLS {
		if config.TLSCertificateFile == "" && config.TLSPrivateKeyFile == "" {
			// When using ACME
			if config.AcmeEndpoint == "" {
				return fmt.Errorf("TLS_ACME_ENDPOINT should be defined when using TLS without providing certificates")
			}
			if config.AcmeEmail == "" {
				return fmt.Errorf("TLS_ACME_EMAIL should be defined when using TLS without providing certificates")
			}
			if config.TLSCertificatesDirectory == "" {
				return fmt.Errorf("TLS_CERTIFICATES_DIRECTORY should be defined when using TLS without providing certificates")
			}
			if config.AcmeChallenge == "" {
				return fmt.Errorf("TLS_ACME_CHALLENGE should be defined (either HTTP or DNS)")
			}
			if config.AcmeChallenge == AcmeChallengeDNS && config.AcmeDNSProvider == "" {
				return fmt.Errorf("TLS_ACME_DNS_PROVIDER shouldn't be empty when using DNS challenge")
			}
		}
		// When have user defined certificates
		if (config.TLSCertificateFile != "" && config.TLSPrivateKeyFile == "") || (config.TLSCertificateFile == "" && config.TLSPrivateKeyFile != "") {
			return fmt.Errorf("both TLS_CERTIFICATE_FILE and TLS_PRIVATE_KEY_FILE must be defined when using custom TLS certificate")
		}
	}

	// HTTP server & api
	if config.HTTPAddress == "" {
		return fmt.Errorf("HTTP_ADDRESS cannot be empty")
	}
	if config.Secret == "" {
		return fmt.Errorf("SECRET cannot be empty")
	}

	// Metrics
	if config.MetricsAddress == "" {
		return fmt.Errorf("METRICS_ADDRESS cannot be empty")
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

	blacklist, err := helpers.NewBlacklist(config.BlacklistURL)
	if err != nil {
		log.Warnln("couldn't create Nixspam Blacklist instance: ", err)
	}

	return &server.Config{
		Hostname:    config.Hostname,
		Ip:          host,
		Port:        uint32(portInt),
		Blacklist:   blacklist,
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

	blacklist, err := helpers.NewBlacklist(config.BlacklistURL)
	if err != nil {
		log.Warnln("couldn't create Nixspam Blacklist instance: ", err)
	}

	return &server.Config{
		Hostname:    config.Hostname,
		Ip:          host,
		Port:        uint32(portInt),
		Blacklist:   blacklist,
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
