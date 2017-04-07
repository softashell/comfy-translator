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
	DB *bolt.DB
}

func NewCache() *Cache {
	db, err := bolt.Open("translation.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}

		return nil
	})
	check(err)

	cache := &Cache{DB: db}

	return cache
}

func (c *Cache) Put(text string, translation string) error {
	err := c.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(text), []byte(translation))

		return err
	})

	return err
}

func (c *Cache) Get(text string) (bool, string) {
	var translation string
	var found bool

	start := time.Now()

	// retrieve the data
	err := c.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			log.Fatalf("Bucket %q not found!", bucketName)
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
	check(err)

	return found, translation
}
