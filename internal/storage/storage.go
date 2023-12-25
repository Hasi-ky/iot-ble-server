package storage

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	cmap "github.com/streamrail/concurrent-map"

	"iot-ble-server/global/globalmemo"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/internal/config"
)

func Setup(ctx context.Context, c config.Config) error {
	log.Info("<storage Setup>: setting up storage package")

	if c.General.UseRedis {
		log.Info("<storage Setup>: setting up Redis client")
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
		go redisKeepAlive(ctx)
	}

	log.Info("<storage Setup>: connecting to PostgreSQL database")
	d, err := sqlx.Open("postgres", c.PostgreSQL.DSN)
	SetDB(&DBLogger{DB: d})
	if err != nil {
		return errors.Wrap(err, "<storage Setup>: PostgreSQL connection error")
	}
	d.SetMaxOpenConns(c.PostgreSQL.MaxOpenConnections)
	d.SetMaxIdleConns(c.PostgreSQL.MaxIdleConnections)
	go pgsqlKeepAlive(ctx, d)
	globalredis.RedisCache = RedisClient()
	globalmemo.MemoCacheDev = cmap.New()
	globalmemo.MemoCacheGw = cmap.New()
	createTables()
	return nil
}

func redisKeepAlive(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				log.WithError(err).Warning("<redisKeepAlive>: ping Redis error, will retry in 30s")
			} else {
				log.Info("<redisKeepAlive>: ping Redis success")
			}
		case <-ctx.Done():
			return
		}
	}
}

func pgsqlKeepAlive(ctx context.Context, db *sqlx.DB) {
	log.Info("psql keep alive")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := db.Ping(); err != nil {
				log.WithError(err).Warning("<pgsqlKeepAlive>: ping PostgreSQL database error, will retry in 30s")
			} else {
				log.Info("<pgsqlKeepAlive>: ping PostgreSQL database success")
			}
		case <-ctx.Done():
			return
		}
	}
}
