package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"bytes"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

type Cache struct {
	db *bolt.DB

	meta cacheMetadata
}

type cacheItem struct {
	Translation string `json:"t"`
	Error       string `json:"e,omitempty"`
	Timestamp   int64  `json:"u,omitempty"`
}

func NewCache(filePath string) (*Cache, error) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	cache := &Cache{db: db}

	cache.readMetadata()

	log.Info("Cache version ", cache.meta.Version)

	cache.migrateDatabase()

	cache.writeMetadata()

	return cache, nil
}

func (c *Cache) Close() error {
	return c.db.Close()
}

func (c *Cache) Put(bucketName, text, translation, cerr string) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket doesn't exist??")
		}

		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(cacheItem{
			Translation: translation,
			Error:       cerr,
			Timestamp:   time.Now().UTC().Unix(),
		}); err != nil {
			return err
		}

		return b.Put([]byte(text), buf.Bytes())
	})

	return err
}

func (c *Cache) Get(bucketName, text string) (string, bool, error) {
	var i cacheItem
	var found bool

	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			log.Warnf("Bucket %q not found! (Probably nothing has been saved to it yet)", bucketName)
			return nil
		}

		val := b.Get([]byte(text))
		if val == nil {
			return nil
		}

		buf := bytes.NewReader(val)
		if err := json.NewDecoder(buf).Decode(&i); err != nil {
			// try to ignore decoding error and just act like we found nothing
			log.Errorf("%v : %s", err, string(val))
			return nil
		}

		found = true

		return nil
	})

	if err != nil {
		log.Fatal(err)
		return "", false, err
	}

	if !found {
		return "", false, nil
	}

	if i.Error != "" {
		if time.Since(time.Unix(i.Timestamp, 0)) > 12*time.Hour {
			return "", false, nil
		}

		return "", true, fmt.Errorf("%s", i.Error)
	}

	return i.Translation, found, nil
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
