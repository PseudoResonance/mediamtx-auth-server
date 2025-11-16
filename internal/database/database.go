package database

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	ttlcache "github.com/jellydator/ttlcache/v3"
	"github.com/pseudoresonance/authserver/internal/config"
)

const TargetSchemaVersion = "2025-11-16T01:44:52+00:00"

/*
MediaMTX passed auth credentials
*/
type Credentials struct {
	Action     string
	Path       string
	QueryToken string
}

type DatabaseManager struct {
	conf   *config.MainConfig
	pool   *pgxpool.Pool
	poller *DatabasePoller

	httpClient http.Client

	cache       *ttlcache.Cache[Credentials, *CredentialData]
	connections *ttlcache.Cache[string, ConnectionRecord]
}

func (d *DatabaseManager) Init(config *config.MainConfig) {
	d.conf = config
	d.poller = &DatabasePoller{db: d, interval: time.Duration(d.conf.Database.PollInterval) * time.Second}

	d.cache = ttlcache.New(
		ttlcache.WithTTL[Credentials, *CredentialData](time.Duration(d.conf.Database.CacheDuration)*time.Second),
		ttlcache.WithDisableTouchOnHit[Credentials, *CredentialData](),
	)
	d.cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[Credentials, *CredentialData]) {
		if credData := item.Value(); credData != nil && credData.Valid {
			// If there are still connections open, check if the creds are still valid before disconnecting them
			if len(credData.connections) > 0 {
				creds := item.Key()
				val, err := d.validateAuth(&creds)
				if err != nil {
					log.Printf("Error while validating auth\n%v\n", err)
					d.revoke(credData)
					return
				}
				if !val {
					d.revoke(credData)
					return
				}
				d.cache.Set(creds, credData, ttlcache.DefaultTTL)
			}
		}
	})

	d.connections = ttlcache.New(
		ttlcache.WithTTL[string, ConnectionRecord](time.Duration(d.conf.Database.ConnectionTrackDuration)*time.Minute),
		ttlcache.WithDisableTouchOnHit[string, ConnectionRecord](),
	)
	d.connections.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[string, ConnectionRecord]) {
		record := item.Value()
		valid := d.validateConnection(record.Info)
		if valid {
			// Reset cache if still valid
			d.connections.Set(item.Key(), record, ttlcache.DefaultTTL)
		}
	})

	d.poller.Start()
	go d.cache.Start()
	go d.connections.Start()

	pgConf, err := pgxpool.ParseConfig("")
	if err != nil {
		log.Fatalf("Error creating PostgreSQL config\n%v\n", err)
	}
	pgConf.ConnConfig.Host = d.conf.Database.Hostname
	pgConf.ConnConfig.Port = uint16(d.conf.Database.Port)
	pgConf.ConnConfig.Database = d.conf.Database.Database
	pgConf.ConnConfig.User = d.conf.Database.Username
	pgConf.ConnConfig.Password = d.conf.Database.Password

	d.pool, err = pgxpool.NewWithConfig(context.Background(), pgConf)
	if err != nil {
		log.Fatalf("Error creating PostgreSQL connection pool\n%v\n", err)
	}
	d.checkSchema()
}

func (d *DatabaseManager) checkSchema() {
	var schemaVersion time.Time
	err := d.pool.QueryRow(context.Background(), `SELECT version FROM versions WHERE application = 'db_version'`).Scan(&schemaVersion)
	if err != nil {
		log.Fatalf("Error fetching PostgreSQL schema version\n%v\n", err)
	}
	targetVersion, err := time.Parse(time.RFC3339, TargetSchemaVersion)
	if err != nil {
		log.Fatalf("Invalid PostgreSQL schema version provided in code! (%v)\n%v\n", TargetSchemaVersion, err)
	}
	log.Printf("Database schema %v\n", schemaVersion)
	if targetVersion.Compare(schemaVersion) > 0 {
		log.Fatalf("Outdated PostgreSQL schema (%v) needs (%v)\n", schemaVersion, targetVersion)
	}
}

func (d *DatabaseManager) Close() {
	d.poller.Close()
	d.cache.Stop()
	d.connections.Stop()
	d.pool.Close()
	d.httpClient.CloseIdleConnections()
}
