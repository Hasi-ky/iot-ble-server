package config

import (
	"time"
)

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogFile       string `mapstructure:"log_file"`
		LogLevel      int    `mapstructure:"log_level"`
		UseRedis      bool   `mapstructure:"use_redis"`
		LocalHost     string `mapstructure:"local_host"`
		BindPort      int    `mapstructure:"bind_port"`
		KeepAliveTime int    `mapstructure:"keepalive_time"`
	} `mapstructure:"general"`

	PostgreSQL struct {
		DSN                string `mapstructure:"dsn"`
		Automigrate        bool
		MaxOpenConnections int `mapstructure:"max_open_connections"`
		MaxIdleConnections int `mapstructure:"max_idle_connections"`
	} `mapstructure:"postgresql"`

	Redis struct {
		URL        string   `mapstructure:"url"` // deprecated
		Servers    []string `mapstructure:"servers"`
		Cluster    bool     `mapstructure:"cluster"`
		MasterName string   `mapstructure:"master_name"`
		PoolSize   int      `mapstructure:"pool_size"`
		Password   string   `mapstructure:"password"`
		Database   int      `mapstructure:"database"`
		TLSEnabled bool     `mapstructure:"tls_enabled"`
		KeyPrefix  string   `mapstructure:"key_prefix"`
	} `mapstructure:"redis"`

	MQTTConfig struct {
		Username             string        `mapstructure:"username"`
		Password             string        `mapstructure:"password"`
		MaxReconnectInterval time.Duration `mapstructure:"max_reconnect_interval"`
		QOS                  uint8         `mapstructure:"qos"`
		CleanSession         bool          `mapstructure:"clean_session"`
		ClientID             string        `mapstructure:"client_id"`
		CACert               string        `mapstructure:"ca_cert"`
		TLSCert              string        `mapstructure:"tls_cert"`
		TLSKey               string        `mapstructure:"tls_key"`
		EventTopicTemplate   string        `mapstructure:"event_topic_template"`
		CommandTopicTemplate string        `mapstructure:"command_topic_template"`
		RetainEvents         bool          `mapstructure:"retain_events"`

		// Topics
		NetworkInTopic    string `mapstructure:"network_in_topic"`
		NetworkInAckTopic string `mapstructure:"network_in_ack_topic"`
		TelemetryTopic    string `mapstructure:"telemetry_topic"`
		RpcTopicTemplate  string `mapstructure:"rpc_topic"`
	}
}

// C holds the global configuration.
var C Config

// Get returns the configuration.
func Get() *Config {
	return &C
}

// Set sets the configuration.
func Set(c Config) {
	C = c
}
