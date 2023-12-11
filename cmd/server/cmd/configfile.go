package cmd

import (
	"os"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"iot-ble-server/internal/config"
)

// when updating this template, don't forget to update config.md!
const configTemplate = `[general]
# Log level
#
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level={{ .General.LogLevel }}

# PostgreSQL settings.
[postgresql]
# PostgreSQL dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable).
#
# Besides using an URL (e.g. 'postgres://user:password@hostname/database?sslmode=disable')
# it is also possible to use the following format:
# 'user=iotware dbname=iotware sslmode=disable'.
#
# The following connection parameters are supported:
#
# * dbname - The name of the database to connect to
# * user - The user to sign in as
# * password - The user's password
# * host - The host to connect to. Values that start with / are for unix domain sockets. (default is localhost)
# * port - The port to bind to. (default is 5432)
# * sslmode - Whether or not to use SSL (default is require, this is not the default for libpq)
# * fallback_application_name - An application_name to fall back to if one isn't provided.
# * connect_timeout - Maximum wait for connection, in seconds. Zero or not specified means wait indefinitely.
# * sslcert - Cert file location. The file must contain PEM encoded data.
# * sslkey - Key file location. The file must contain PEM encoded data.
# * sslrootcert - The location of the root certificate file. The file must contain PEM encoded data.
#
# Valid values for sslmode are:
#
# * disable - No SSL
# * require - Always SSL (skip verification)
# * verify-ca - Always SSL (verify that the certificate presented by the server was signed by a trusted CA)
# * verify-full - Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)
dsn="{{ .PostgreSQL.DSN }}"

# Automatically apply database migrations.

# Max open connections.
#
# This sets the max. number of open connections that are allowed in the
# PostgreSQL connection pool (0 = unlimited).
max_open_connections={{ .PostgreSQL.MaxOpenConnections }}

# Max idle connections.
#
# This sets the max. number of idle connections in the PostgreSQL connection
# pool (0 = no idle connections are retained).
max_idle_connections={{ .PostgreSQL.MaxIdleConnections }}


# Redis settings
[redis]

# Server address or addresses.
#
# Set multiple addresses when connecting to a cluster.
servers=[{{ range $index, $elm := .Redis.Servers }}
  "{{ $elm }}",{{ end }}
]

# Password.
#
# Set the password when connecting to Redis requires password authentication.
password="{{ .Redis.Password }}"

# Database index.
#
# By default, this can be a number between 0-15.
database={{ .Redis.Database }}


# Redis Cluster.
#
# Set this to true when the provided URLs are pointing to a Redis Cluster
# instance.
cluster={{ .Redis.Cluster }}

# Master name.
#
# Set the master name when the provided URLs are pointing to a Redis Sentinel
# instance.
master_name="{{ .Redis.MasterName }}"

# Connection pool size.
#
# Default (when set to 0) is 10 connections per every CPU.
pool_size={{ .Redis.PoolSize }}

# TLS enabled.
#
# Note: tis will enable TLS, but it will not validate the certificate
# used by the server.
tls_enabled={{ .Redis.TLSEnabled }}

# Key prefix.
#
# A key prefix can be used to avoid key collisions when multiple deployments
# are using the same Redis database and it is not possible to separate
# keys by database index (e.g. when using Redis Cluster, which does not
# support multiple databases).
key_prefix="{{ .Redis.KeyPrefix }}"
`

var configCmd = &cobra.Command{
	Use:   "configfile",
	Short: "Print the IOT BLE Server configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := template.Must(template.New("config").Parse(configTemplate))
		err := t.Execute(os.Stdout, config.C)
		if err != nil {
			return errors.Wrap(err, "execute config template error")
		}
		return nil
	},
}
