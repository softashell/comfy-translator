package cache

import (
	log "github.com/Sirupsen/logrus"
)

type Metadata struct {
	ID          int `storm:"id"`
	Version     int
	LastCleanup int64
}

func (c *Cache) readMetadata() error {
	err := c.db.One("ID", 1, &c.meta)

	if err != nil {
		log.Error("failed to get metadata:", err)
	}

	return err
}

func (c *Cache) writeMetadata() error {
	if c.meta.ID == 0 {
		c.meta.ID = 1
	}

	err := c.db.Save(&c.meta)
	if err != nil {
		log.Error("failed to update metadata:", err)
	}

	return err
}
