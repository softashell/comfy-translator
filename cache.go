package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

var (
	bucketName = []byte("translations")
)

type Cache struct {
	database *bolt.DB
}

func NewCache() (*Cache, error) {
	db, err := bolt.Open("translation.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	cache := &Cache{database: db}

	return cache, nil
}

func (c *Cache) Close() error {
	return c.database.Close()
}

func (c *Cache) Put(bucketName, text, translation string) error {
	err := c.database.Update(func(tx *bolt.Tx) error {
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
	err := c.database.View(func(tx *bolt.Tx) error {
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
