package cache

import (
	"bytes"
	"time"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
	bolt "github.com/coreos/bbolt"

	"gitgud.io/softashell/comfy-translator/translator/google"
)

const latestVersion = 3

func (c *Cache) migrateDatabase() error {
	if c.meta.Version == latestVersion {
		return nil
	}

	log.Infof("Need to run %d database migrations", latestVersion-c.meta.Version)

	for c.meta.Version < latestVersion {
		c.migrate(c.meta.Version)
	}

	return nil
}

func (c *Cache) migrate(ver int) {
	start := time.Now()

	tgt := ver + 1

	log.Info("Running migration ", tgt)

	var err error

	switch tgt {
	case 1:
		err = c.migration1()
	case 2:
		err = c.migration2()
	case 3:
		err = c.migration3()
	}

	log := log.WithFields(log.Fields{
		"time": time.Since(start),
		"ver":  ver,
		"tgt":  tgt,
	})

	if err != nil {
		log.Fatal("Migration failed: ", err)
	}

	c.meta.Version = tgt

	log.Info("Finished migration")
}

func (c *Cache) migration1() error {
	// Move old translations to Google bucket

	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("translations"))
		if b == nil {
			return nil
		}

		googleBucket, err := tx.CreateBucketIfNotExists([]byte("Google"))
		if err != nil {
			return err
		}

		cur := b.Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			if err := googleBucket.Put(k, v); err != nil {
				log.Error(err)
				return err
			}
		}

		if err := tx.DeleteBucket([]byte("translations")); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (c *Cache) migration2() error {
	// Use json and store errors and timestamp

	buckets, err := c.getBuckets()
	if err != nil {
		return err
	}

	timestamp := time.Now().UTC().Unix()
	err = c.db.Update(func(tx *bolt.Tx) error {
		for _, bucketName := range buckets {
			b := tx.Bucket(bucketName)
			if b == nil {
				return nil
			}

			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if v == nil {
					continue
				}

				buf := bytes.NewBuffer(nil)
				if err := json.NewEncoder(buf).Encode(cacheItem{
					Translation: string(v),
					Timestamp:   timestamp,
				}); err != nil {
					return err
				}

				if err := b.Put(k, buf.Bytes()); err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

func (c *Cache) migration3() error {
	// clean google cache

	var i cacheItem

	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Google"))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				continue
			}

			buf := bytes.NewReader(v)
			if err := json.NewDecoder(buf).Decode(&i); err != nil {
				log.Errorf("%v : %s", err, string(v))
				b.Delete(k)
				continue
			}

			if len(i.Translation) > 0 && google.IsTranslationGarbage(i.Translation) {
				log.Debugf("Removing: %q => %q", string(k), i.Translation)
				b.Delete(k)
			}
		}

		return nil
	})

	return err
}

func (c *Cache) getBuckets() ([][]byte, error) {
	var buckets [][]byte

	err := c.db.View(func(tx *bolt.Tx) error {
		if err := tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			if string(name) != metadataName {
				buckets = append(buckets, name)
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	return buckets, err
}
