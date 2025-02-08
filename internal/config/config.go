package config

type Config struct {
	Global  GlobalConfig  `toml:"global" validate:"required"`
	Log     LogConfig     `toml:"log" validate:"required"`
	Servers ServersConfig `toml:"servers" validate:"required"`
	Sentry SentryConfig `toml:"sentry" validate:"required"`
}

type GlobalConfig struct {
	Env string `toml:"env" validate:"required,oneof=dev stage prod"`
}

type LogConfig struct {
	Level string `toml:"level" validate:"required,oneof=debug info warn error"`
}

type ServersConfig struct {
	Debug DebugServerConfig `toml:"debug" validate:"required"`
}

type SentryConfig struct {
    DSN string `toml:"dsn" validate:"required"`
}

type DebugServerConfig struct {
	Addr string `toml:"addr" validate:"required,hostname_port"`
}
