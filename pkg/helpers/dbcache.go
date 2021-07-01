// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package helpers

import (
	"database/sql"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	// maxIdleTimeSeconds is how long we'll keep an idle connection
	// open in the pool.
	maxIdleTimeSeconds = 5 * 60
)

// NewDBCache returns a new empty cache of DB connections.
func NewDBCache(driverName string) *DBCache {
	return &DBCache{
		driverName: driverName,
		items:      make(map[string]*sql.DB),
	}
}

// DbCache keeps a cache of existing *sql.DBs for a specific driver
// for different data sources. It avoids us creating a new one
// everytime we run commands against a database.
type DBCache struct {
	driverName string
	mutex      sync.Mutex
	items      map[string]*sql.DB
}

// Get returns an existing DB for sourceName if one exists. If not it
// creates a new one and returns it (after storing it in the
// cache). Safe to call from different goroutines.
func (c *DBCache) Get(sourceName string) (*sql.DB, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	db, found := c.items[sourceName]
	if !found {
		newDb, err := sql.Open(c.driverName, sourceName)
		if err != nil {
			return nil, errors.Wrap(err, "opening new DB")
		}
		db = newDb
		c.items[sourceName] = newDb
		db.SetConnMaxIdleTime(time.Second * maxIdleTimeSeconds)
	}
	return db, nil
}

// TODO: provide a way to remove from the cache when we delete a
// database. Without this, when a database is removed the connections
// will eventually be removed once the idle timeout expires, but the
// cached sql.DB won't ever be removed from the map.
