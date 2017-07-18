package cache

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

type Cache struct {
	db *bolt.DB
}

func NewCache(filePath string) (*Cache, error) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	cache := &Cache{db: db}

	err = cache.CreateBucket("translations")
	if err != nil {
		log.Error("Failed to create missing bucket \"translations\"")
	}

	return cache, nil
}

func (c *Cache) Close() error {
	return c.db.Close()
}

func (c *Cache) Put(bucketName, text, translation string) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(text), []byte(translation))

		return err
	})

	return err
}

func (c *Cache) Get(bucketName, text string) (bool, string) {
	var translation string
	var found bool

	start := time.Now()

	// retrieve the data
	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			log.Warnf("Bucket %q not found! (Probably nothing has been saved to it yet)", bucketName)
			return nil
		}

		val := bucket.Get([]byte(text))

		if val != nil {
			found = true
			translation = string(val)

			log.WithFields(log.Fields{
				"time": time.Since(start),
			}).Debugf("Cache: %q", translation)
		}

		return nil
	})

	if err != nil {
		log.Error(err)
	}

	return found, translation
}

func (c *Cache) CreateBucket(bucketName string) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
