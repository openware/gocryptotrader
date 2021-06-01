package config

import (
	"sync"
	"time"

	"github.com/openware/gocryptotrader/database"
	gctscript "github.com/openware/gocryptotrader/gctscript/vm"
	irixCfg "github.com/openware/irix/config"
	"github.com/openware/irix/portfolio"
	"github.com/openware/irix/portfolio/banking"
	"github.com/openware/pkg/currency"
	"github.com/openware/pkg/log"
)

// Constants declared here are filename strings and test strings
const (
	FXProviderFixer                      = "fixer"
	EncryptedFile                        = "config.dat"
	File                                 = "config.json"
	TestFile                             = "../testdata/configtest.json"
	fileEncryptionPrompt                 = 0
	fileEncryptionEnabled                = 1
	fileEncryptionDisabled               = -1
	pairsLastUpdatedWarningThreshold     = 30 // 30 days
	defaultHTTPTimeout                   = time.Second * 15
	defaultWebsocketResponseCheckTimeout = time.Millisecond * 30
	defaultWebsocketResponseMaxLimit     = time.Second * 7
	defaultWebsocketOrderbookBufferLimit = 5
	defaultWebsocketTrafficTimeout       = time.Second * 30
	maxAuthFailures                      = 3
	defaultNTPAllowedDifference          = 50000000
	defaultNTPAllowedNegativeDifference  = 50000000
	DefaultAPIKey                        = "Key"
	DefaultAPISecret                     = "Secret"
	DefaultAPIClientID                   = "ClientID"
)

// Constants here hold some messages
const (
	ErrExchangeNameEmpty                       = "exchange #%d name is empty"
	ErrExchangeAvailablePairsEmpty             = "exchange %s available pairs is empty"
	ErrExchangeEnabledPairsEmpty               = "exchange %s enabled pairs is empty"
	ErrExchangeBaseCurrenciesEmpty             = "exchange %s base currencies is empty"
	ErrExchangeNotFound                        = "exchange %s not found"
	ErrNoEnabledExchanges                      = "no exchanges enabled"
	ErrCryptocurrenciesEmpty                   = "cryptocurrencies variable is empty"
	ErrFailureOpeningConfig                    = "fatal error opening %s file. Error: %s"
	ErrCheckingConfigValues                    = "fatal error checking config values. Error: %s"
	ErrSavingConfigBytesMismatch               = "config file %q bytes comparison doesn't match, read %s expected %s"
	WarningWebserverCredentialValuesEmpty      = "webserver support disabled due to empty Username/Password values"
	WarningWebserverListenAddressInvalid       = "webserver support disabled due to invalid listen address"
	WarningExchangeAuthAPIDefaultOrEmptyValues = "exchange %s authenticated API support disabled due to default/empty APIKey/Secret/ClientID values"
	WarningPairsLastUpdatedThresholdExceeded   = "exchange %s last manual update of available currency pairs has exceeded %d days. Manual update required!"
)

// Constants here define unset default values displayed in the config.json
// file
const (
	APIURLNonDefaultMessage              = "NON_DEFAULT_HTTP_LINK_TO_EXCHANGE_API"
	WebsocketURLNonDefaultMessage        = "NON_DEFAULT_HTTP_LINK_TO_WEBSOCKET_EXCHANGE_API"
	DefaultUnsetAPIKey                   = "Key"
	DefaultUnsetAPISecret                = "Secret"
	DefaultUnsetAccountPlan              = "accountPlan"
	DefaultForexProviderExchangeRatesAPI = "ExchangeRateHost"
)

// Variables here are used for configuration
var (
	Cfg Config
	m   sync.Mutex
)

// Config is the overarching object that holds all the information for
// prestart management of Portfolio, Communications, Webserver and Enabled
// Exchanges
type Config struct {
	Name              string                   `json:"name"`
	DataDirectory     string                   `json:"dataDirectory"`
	EncryptConfig     int                      `json:"encryptConfig"`
	GlobalHTTPTimeout time.Duration            `json:"globalHTTPTimeout"`
	Database          database.Config          `json:"database"`
	Logging           log.Config               `json:"logging"`
	ConnectionMonitor ConnectionMonitorConfig  `json:"connectionMonitor"`
	Profiler          Profiler                 `json:"profiler"`
	NTPClient         NTPClientConfig          `json:"ntpclient"`
	GCTScript         gctscript.Config         `json:"gctscript"`
	Currency          CurrencyConfig           `json:"currencyConfig"`
	Communications    CommunicationsConfig     `json:"communications"`
	RemoteControl     RemoteControlConfig      `json:"remoteControl"`
	Portfolio         portfolio.Base           `json:"portfolioAddresses"`
	Exchanges         []irixCfg.ExchangeConfig `json:"exchanges"`
	BankAccounts      []banking.Account        `json:"bankAccounts"`

	// Deprecated config settings, will be removed at a future date
	Webserver           *WebserverConfig          `json:"webserver,omitempty"`
	CurrencyPairFormat  *CurrencyPairFormatConfig `json:"currencyPairFormat,omitempty"`
	FiatDisplayCurrency *currency.Code            `json:"fiatDispayCurrency,omitempty"`
	Cryptocurrencies    *currency.Currencies      `json:"cryptocurrencies,omitempty"`
	SMS                 *SMSGlobalConfig          `json:"smsGlobal,omitempty"`
	// encryption session values
	storedSalt []byte
	sessionDK  []byte
}

// ConnectionMonitorConfig defines the connection monitor variables to ensure
// that there is internet connectivity
type ConnectionMonitorConfig struct {
	DNSList          []string      `json:"preferredDNSList"`
	PublicDomainList []string      `json:"preferredDomainList"`
	CheckInterval    time.Duration `json:"checkInterval"`
}

// Profiler defines the profiler configuration to enable pprof
type Profiler struct {
	Enabled              bool `json:"enabled"`
	MutexProfileFraction int  `json:"mutex_profile_fraction"`
}

// NTPClientConfig defines a network time protocol configuration to allow for
// positive and negative differences
type NTPClientConfig struct {
	Level                     int            `json:"enabled"`
	Pool                      []string       `json:"pool"`
	AllowedDifference         *time.Duration `json:"allowedDifference"`
	AllowedNegativeDifference *time.Duration `json:"allowedNegativeDifference"`
}

// GRPCConfig stores the gRPC settings
type GRPCConfig struct {
	Enabled                bool   `json:"enabled"`
	ListenAddress          string `json:"listenAddress"`
	GRPCProxyEnabled       bool   `json:"grpcProxyEnabled"`
	GRPCProxyListenAddress string `json:"grpcProxyListenAddress"`
}

// DepcrecatedRPCConfig stores the deprecatedRPCConfig settings
type DepcrecatedRPCConfig struct {
	Enabled       bool   `json:"enabled"`
	ListenAddress string `json:"listenAddress"`
}

// WebsocketRPCConfig stores the websocket config info
type WebsocketRPCConfig struct {
	Enabled             bool   `json:"enabled"`
	ListenAddress       string `json:"listenAddress"`
	ConnectionLimit     int    `json:"connectionLimit"`
	MaxAuthFailures     int    `json:"maxAuthFailures"`
	AllowInsecureOrigin bool   `json:"allowInsecureOrigin"`
}

// RemoteControlConfig stores the RPC services config
type RemoteControlConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`

	GRPC          GRPCConfig           `json:"gRPC"`
	DeprecatedRPC DepcrecatedRPCConfig `json:"deprecatedRPC"`
	WebsocketRPC  WebsocketRPCConfig   `json:"websocketRPC"`
}

// WebserverConfig stores the old webserver config
type WebserverConfig struct {
	Enabled                      bool   `json:"enabled"`
	AdminUsername                string `json:"adminUsername"`
	AdminPassword                string `json:"adminPassword"`
	ListenAddress                string `json:"listenAddress"`
	WebsocketConnectionLimit     int    `json:"websocketConnectionLimit"`
	WebsocketMaxAuthFailures     int    `json:"websocketMaxAuthFailures"`
	WebsocketAllowInsecureOrigin bool   `json:"websocketAllowInsecureOrigin"`
}

// Post holds the bot configuration data
type Post struct {
	Data Config `json:"data"`
}

// CurrencyPairFormatConfig stores the users preferred currency pair display
type CurrencyPairFormatConfig struct {
	Uppercase bool   `json:"uppercase"`
	Delimiter string `json:"delimiter,omitempty"`
	Separator string `json:"separator,omitempty"`
	Index     string `json:"index,omitempty"`
}

// BankTransaction defines a related banking transaction
type BankTransaction struct {
	ReferenceNumber     string `json:"referenceNumber"`
	TransactionNumber   string `json:"transactionNumber"`
	PaymentInstructions string `json:"paymentInstructions"`
}

// CurrencyConfig holds all the information needed for currency related manipulation
type CurrencyConfig struct {
	ForexProviders                []currency.FXSettings     `json:"forexProviders"`
	CryptocurrencyProvider        CryptocurrencyProvider    `json:"cryptocurrencyProvider"`
	Cryptocurrencies              currency.Currencies       `json:"cryptocurrencies"`
	CurrencyPairFormat            *CurrencyPairFormatConfig `json:"currencyPairFormat"`
	FiatDisplayCurrency           currency.Code             `json:"fiatDisplayCurrency"`
	CurrencyFileUpdateDuration    time.Duration             `json:"currencyFileUpdateDuration"`
	ForeignExchangeUpdateDuration time.Duration             `json:"foreignExchangeUpdateDuration"`
}

// CryptocurrencyProvider defines coinmarketcap tools
type CryptocurrencyProvider struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Verbose     bool   `json:"verbose"`
	APIkey      string `json:"apiKey"`
	AccountPlan string `json:"accountPlan"`
}

// CommunicationsConfig holds all the information needed for each
// enabled communication package
type CommunicationsConfig struct {
	SlackConfig     SlackConfig     `json:"slack"`
	SMSGlobalConfig SMSGlobalConfig `json:"smsGlobal"`
	SMTPConfig      SMTPConfig      `json:"smtp"`
	TelegramConfig  TelegramConfig  `json:"telegram"`
}

// IsAnyEnabled returns whether or any any comms relayers
// are enabled
func (c *CommunicationsConfig) IsAnyEnabled() bool {
	if c.SMSGlobalConfig.Enabled ||
		c.SMTPConfig.Enabled ||
		c.SlackConfig.Enabled ||
		c.TelegramConfig.Enabled {
		return true
	}
	return false
}

// SlackConfig holds all variables to start and run the Slack package
type SlackConfig struct {
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	Verbose           bool   `json:"verbose"`
	TargetChannel     string `json:"targetChannel"`
	VerificationToken string `json:"verificationToken"`
}

// SMSContact stores the SMS contact info
type SMSContact struct {
	Name    string `json:"name"`
	Number  string `json:"number"`
	Enabled bool   `json:"enabled"`
}

// SMSGlobalConfig structure holds all the variables you need for instant
// messaging and broadcast used by SMSGlobal
type SMSGlobalConfig struct {
	Name     string       `json:"name"`
	From     string       `json:"from"`
	Enabled  bool         `json:"enabled"`
	Verbose  bool         `json:"verbose"`
	Username string       `json:"username"`
	Password string       `json:"password"`
	Contacts []SMSContact `json:"contacts"`
}

// SMTPConfig holds all variables to start and run the SMTP package
type SMTPConfig struct {
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	Verbose         bool   `json:"verbose"`
	Host            string `json:"host"`
	Port            string `json:"port"`
	AccountName     string `json:"accountName"`
	AccountPassword string `json:"accountPassword"`
	From            string `json:"from"`
	RecipientList   string `json:"recipientList"`
}

// TelegramConfig holds all variables to start and run the Telegram package
type TelegramConfig struct {
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	Verbose           bool   `json:"verbose"`
	VerificationToken string `json:"verificationToken"`
}
