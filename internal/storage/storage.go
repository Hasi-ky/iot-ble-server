package storage

import (
	"context"
	"crypto/tls"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"

	"iot-ble-server/internal/config"
)

func Setup(c config.Config) error {
	log.Info("storage: setting up storage package")

	if c.General.UseRedis {
		log.Info("storage: setting up Redis client")
		if len(c.Redis.Servers) == 0 {
			return errors.New("at least one redis server must be configured")
		}

		var tlsConfig *tls.Config
		if c.Redis.TLSEnabled {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		if c.Redis.Cluster {
			redisClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:     c.Redis.Servers,
				PoolSize:  c.Redis.PoolSize,
				Password:  c.Redis.Password,
				TLSConfig: tlsConfig,
			})
		} else if c.Redis.MasterName != "" {
			redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:       c.Redis.MasterName,
				SentinelAddrs:    c.Redis.Servers,
				SentinelPassword: c.Redis.Password,
				DB:               c.Redis.Database,
				PoolSize:         c.Redis.PoolSize,
				Password:         c.Redis.Password,
				TLSConfig:        tlsConfig,
			})
		} else {
			redisClient = redis.NewClient(&redis.Options{
				Addr:      c.Redis.Servers[0],
				DB:        c.Redis.Database,
				Password:  c.Redis.Password,
				PoolSize:  c.Redis.PoolSize,
				TLSConfig: tlsConfig,
			})
		}

		// Redis keep alive
		go redisKeepAlive()
	}

	log.Info("storage: connecting to PostgreSQL database")
	d, err := sqlx.Open("postgres", c.PostgreSQL.DSN)
	if err != nil {
		return errors.Wrap(err, "storage: PostgreSQL connection error")
	}
	d.SetMaxOpenConns(c.PostgreSQL.MaxOpenConnections)
	d.SetMaxIdleConns(c.PostgreSQL.MaxIdleConnections)
	go pgsqlKeepAlive(d)

	return nil
}

func redisKeepAlive()  {
	log.Info("redis keep alive")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				log.WithError(err).Warning("storage: ping Redis error, will retry in 30s")
			} else {
				log.Info("storage: ping Redis success")
			}
		}
	}
}

func pgsqlKeepAlive(db *sqlx.DB)  {
	log.Info("psql keep alive")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := db.Ping(); err != nil {
				log.WithError(err).Warning("storage: ping PostgreSQL database error, will retry in 30s")
			} else {
				log.Info("storage: ping PostgreSQL database success")
			}
		}
	}
}