// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package helpers

import (
	"database/sql"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

const (
	// maxIdleTimeSeconds is how long we'll keep an idle connection
	// open in the pool.
	maxConnIdleTimeSeconds = 5 * 60

	// dbRemovalSeconds is how long we'll wait after a DB is last used
	// to remove it from the cache.
	dbRemovalSeconds = 20 * 60

	// sweepIntervalSeconds is how long we wait between sweeps.
	sweepIntervalSeconds = 60

	// sweeperStopTimeoutSeconds is how long we'll wait for the
	// sweeper goroutine to stop before complaining.
	sweeperStopTimeoutSeconds = 5
)

// NewDBCache returns a new empty cache of DB connections.
func NewDBCache(driverName string) *DBCache {
	cache := DBCache{
		driverName: driverName,
		items:      make(map[string]*dbEntry),
		stop:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}
	go cache.runSweeper()
	return &cache
}

type dbEntry struct {
	db       *sql.DB
	lastUsed time.Time
}

// DbCache keeps a cache of existing *sql.DBs for a specific driver
// for different data sources. It avoids us creating a new one
// everytime we run commands against a database.
type DBCache struct {
	driverName string
	mutex      sync.Mutex
	items      map[string]*dbEntry
	stop       chan struct{}
	stopped    chan struct{}
}

// Get returns an existing DB for sourceName if one exists. If not it
// creates a new one and returns it (after storing it in the
// cache). Safe to call from different goroutines.
func (c *DBCache) Get(sourceName string) (*sql.DB, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, found := c.items[sourceName]
	if !found {
		newDB, err := sql.Open(c.driverName, sourceName)
		if err != nil {
			return nil, errors.Wrap(err, "opening new DB")
		}
		newDB.SetConnMaxIdleTime(maxConnIdleTimeSeconds * time.Second)
		entry = &dbEntry{
			db:       newDB,
			lastUsed: time.Now(),
		}
		c.items[sourceName] = entry
	} else {
		entry.lastUsed = time.Now()
	}
	return entry.db, nil
}

func (c *DBCache) runSweeper() {
	defer close(c.stopped)
	ticker := time.NewTicker(sweepIntervalSeconds * time.Second)
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.sweep()
		}
	}
}

func (c *DBCache) sweep() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.items {
		removeTime := entry.lastUsed.Add(dbRemovalSeconds * time.Second)
		if now.Before(removeTime) {
			continue
		}
		delete(c.items, key)
		// TODO: logging here?
		entry.db.Close()
	}
}

func (c *DBCache) stopSweeper() error {
	close(c.stop)
	select {
	case <-time.After(sweeperStopTimeoutSeconds * time.Second):
		return errors.Errorf("timed out waiting for sweeper to stop")
	case <-c.stopped:
		return nil
	}
}

func (c *DBCache) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var errs *multierror.Error
	for _, entry := range c.items {
		errs = multierror.Append(errs, entry.db.Close())
	}

	// Stop the sweeper goroutine and wait for it to finish.
	errs = multierror.Append(errs, c.stopSweeper())
	return errs.ErrorOrNil()
}
