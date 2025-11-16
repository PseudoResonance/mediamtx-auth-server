package database

import (
	"context"

	"github.com/jellydator/ttlcache/v3"
)

/*
Validate credentials against the cache and database and handle new connections
*/
func (d *DatabaseManager) ValidateAuth(req *Credentials, connection *Connection) (bool, error) {
	// Check cache
	if cacheVal := d.cache.Get(*req); cacheVal != nil {
		credData := cacheVal.Value()
		if credData == nil {
			return false, nil
		}
		if credData.Valid {
			defer d.registerConnection(req, credData, connection)
		}
		return credData.Valid, nil
	}

	// Check database
	valid, err := d.validateAuth(req)
	if err != nil {
		return false, err
	}
	credData := CredentialData{Valid: valid}
	defer d.cache.Set(*req, &credData, ttlcache.DefaultTTL)
	if valid {
		defer d.registerConnection(req, &credData, connection)
	}
	return valid, nil
}

/*
Internal function to validate credentials against the database
*/
func (d *DatabaseManager) validateAuth(req *Credentials) (bool, error) {
	rows, err := d.pool.Query(context.Background(), "SELECT 1 FROM stream_auth WHERE path = $1 AND action = $2 AND queryToken = $3", req.Path, req.Action, req.QueryToken)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

/*
Register a connection for tracking
*/
func (d *DatabaseManager) registerConnection(req *Credentials, credData *CredentialData, connection *Connection) {
	if connection == nil {
		// One time connection (ex: HLS)
		return
	}
	credData.addConnection(connection.Id)
	d.connections.Set(connection.Id, ConnectionRecord{Creds: req, Info: *connection}, ttlcache.DefaultTTL)
}
