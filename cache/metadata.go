package cache

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	bolt "github.com/coreos/bbolt"
)

const metadataName = "___metadata"

type cacheMetadata struct {
	Version     int
	LastCleanup int64
}

func (c *Cache) readMetadata() error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(metadataName))
		if err != nil {
			return err
		}

		val := b.Get([]byte(metadataName))
		if val != nil {
			json.Unmarshal(val, &c.meta)
		}

		return err
	})
	if err != nil {
		log.Error(err)
	}

	return err
}

func (c *Cache) writeMetadata() error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metadataName))

		bytes, err := json.Marshal(&c.meta)
		if err != nil {
			log.Fatal(err)
		}

		return b.Put([]byte(metadataName), bytes)
	})

	return err
}
