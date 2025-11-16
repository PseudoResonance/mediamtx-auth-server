package database

import (
	"slices"
	"sync"
)

type CredentialData struct {
	mutex       sync.RWMutex
	connections []string
	Valid       bool
}

func (d *CredentialData) addConnection(conn string) {
	d.mutex.Lock()
	d.connections = append(d.connections, conn)
	d.mutex.Unlock()
}

func (d *CredentialData) removeConnection(conn string) {
	d.mutex.Lock()
	i := slices.Index(d.connections, conn)
	if i >= 0 {
		ret := make([]string, len(d.connections)-1)
		ret = append(ret, d.connections[:i]...)
		d.connections = append(ret, d.connections[i+1:]...)
	}
	d.mutex.Unlock()
}

func (d *CredentialData) getAndClearConnections() []string {
	d.mutex.Lock()
	conns := d.connections
	d.connections = []string{}
	d.mutex.Unlock()
	return conns
}
