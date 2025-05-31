package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	Global   GlobalConfig   `toml:"global"`
	Log      LogConfig      `toml:"log"`
	Sentry   SentryConfig   `toml:"sentry"`
	Tracing  TracingConfig  `toml:"tracing"`
	Servers  ServersConfig  `toml:"servers"`
	Stores   StoresConfig   `toml:"stores"`
	Clients  ClientsConfig  `toml:"clients"`
	Services ServicesConfig `toml:"services"`
}

type GlobalConfig struct {
	Name string `toml:"name"`
	Env  string `toml:"env" validate:"oneof=local dev stage prod"`
}

func (c GlobalConfig) IsProduction() bool {
	return c.Env == "prod"
}

type LogConfig struct {
	Level string `toml:"level" validate:"oneof=debug info warn error"`
}

type SentryConfig struct {
	DSN string `toml:"dsn"`
}

type TracingConfig struct {
	Enabled   bool   `toml:"enabled"`
	AgentAddr string `toml:"agent_addr"`
}

type ServersConfig struct {
	Debug   DebugServerConfig   `toml:"debug"`
	Client  ClientServerConfig  `toml:"client"`
	Manager ManagerServerConfig `toml:"manager"`
}

type DebugServerConfig struct {
	Addr string `toml:"addr" validate:"hostname_port"`
}

type ClientServerConfig struct {
	Addr           string               `toml:"addr" validate:"hostname_port"`
	AllowOrigins   []string             `toml:"alloworigins"`
	SecWsProtocol  string               `toml:"secwsprotocol"`
	RequiredAccess RequiredAccessConfig `toml:"requiredaccess"`
}

type ManagerServerConfig struct {
	Addr           string               `toml:"addr" validate:"hostname_port"`
	AllowOrigins   []string             `toml:"alloworigins"`
	SecWsProtocol  string               `toml:"secwsprotocol"`
	RequiredAccess RequiredAccessConfig `toml:"requiredaccess"`
}

type RequiredAccessConfig struct {
	Resource string `toml:"resource"`
	Role     string `toml:"role"`
}

type StoresConfig struct {
	Use    string       `toml:"use"`
	SQLite SQLiteConfig `toml:"sqlite"`
	PSQL   PSQLConfig   `toml:"psql"`
}

type SQLiteConfig struct{}

type PSQLConfig struct {
	Addr     string `toml:"addr"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	Debug    bool   `toml:"debug"`
}

type ClientsConfig struct {
	Keycloak KeycloakConfig `toml:"keycloak"`
}

type KeycloakConfig struct {
	BasePath     string `toml:"basepath"`
	Realm        string `toml:"realm"`
	ClientID     string `toml:"clientid"`
	ClientSecret string `toml:"clientsecret"`
	DebugMode    bool   `toml:"debugmode"`
}

type ServicesConfig struct {
	ManagerLoad          ManagerLoadConfig          `toml:"manager_load"`
	ManagerScheduler     ManagerSchedulerConfig     `toml:"manager_scheduler"`
	MsgProducer          MsgProducerConfig          `toml:"msgproducer"`
	Outbox               OutboxConfig               `toml:"outbox"`
	AFCVerdictsProcessor AFCVerdictsProcessorConfig `toml:"afc_verdicts_processor"`
}

type ManagerLoadConfig struct {
	MaxProblemsAtSameTime int `toml:"max_problems_at_same_time"`
}

type ManagerSchedulerConfig struct {
	Period time.Duration `toml:"period"`
}

type MsgProducerConfig struct {
	Brokers    []string `toml:"brokers"`
	Topic      string   `toml:"topic"`
	BatchSize  int      `toml:"batchsize"`
	EncryptKey string   `toml:"encryptkey"`
}

type OutboxConfig struct {
	Workers    int           `toml:"workers"`
	IdleTime   time.Duration `toml:"idletime"`
	ReserveFor time.Duration `toml:"reservefor"`
}

type AFCVerdictsProcessorConfig struct {
	Brokers                  []string `toml:"brokers"`
	Consumers                int      `toml:"consumers"`
	ConsumerGroup            string   `toml:"consumer_group"`
	BatchSize                int      `toml:"batch_size"`
	VerdictsTopic            string   `toml:"verdicts_topic"`
	VerdictsDLQTopic         string   `toml:"verdicts_dlq_topic"`
	VerdictsSigningPublicKey string   `toml:"verdicts_signing_public_key"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}
