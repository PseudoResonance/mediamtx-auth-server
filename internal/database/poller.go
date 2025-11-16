package database

import (
	"context"
	"log"
	"sync"
	"time"

	ttlcache "github.com/jellydator/ttlcache/v3"
)

type DatabasePoller struct {
	db *DatabaseManager

	mutex    sync.RWMutex
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	lastPoll time.Time
}

func (d *DatabasePoller) Start() {
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.SetLastPoll(time.Now().UTC())
	go d.loop()
}

func (d *DatabasePoller) Close() {
	d.cancel()
}

func (d *DatabasePoller) GetLastPoll() time.Time {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.lastPoll
}

func (d *DatabasePoller) SetLastPoll(timestamp time.Time) {
	d.mutex.Lock()
	d.lastPoll = timestamp
	d.mutex.Unlock()
}

func (d *DatabasePoller) poll() {
	pollTime := d.GetLastPoll()
	d.SetLastPoll(time.Now().UTC())
	// No need to poll if cache is empty
	if d.db.cache.Len() == 0 {
		return
	}
	rows, err := d.db.pool.Query(context.Background(), "SELECT path, action, queryToken FROM stream_auth WHERE created_at > $1", pollTime)
	if err != nil {
		log.Printf("Error while polling database\n%v\n", err)
		return
	}
	defer rows.Close()
	var testCreds Credentials
	for rows.Next() {
		err := rows.Scan(&testCreds.Path, &testCreds.Action, &testCreds.QueryToken)
		if err != nil {
			log.Printf("Error while parsing database column\n%v\n", err)
			continue
		}
		if cacheItem := d.db.cache.Get(testCreds); cacheItem != nil {
			credData := cacheItem.Value()
			if credData == nil {
				credData = &CredentialData{}
			}
			if !credData.Valid {
				credData.Valid = true
				d.db.cache.Set(testCreds, credData, ttlcache.DefaultTTL)
			}
		}
	}
}

func (d *DatabasePoller) loop() {
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			d.poll()
		}
		time.Sleep(d.interval)
	}
}
