package cache

import (
	"fmt"
	"time"

	"go.etcd.io/bbolt"

	log "github.com/Sirupsen/logrus"
)

type oldCacheItem struct {
	Translation string           `json:"t"`
	ErrorCode   translationError `json:"c,omitempty"`
	ErrorText   string           `json:"e,omitempty"`
	Timestamp   int64            `json:"u,omitempty"`
}

const latestVersion = 1

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
	// Move raw boltdb to storm structure

	buckets := [][]byte{
		[]byte("Bing"),
		[]byte("Yandex"),
		[]byte("Google"),
	}

	for _, bucketName := range buckets {
		count := 0

		log.Infof("migrating %s", string(bucketName))

		items := []Item{}

		err := c.db.Bolt.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				return fmt.Errorf("bucket doesn't exist: %s", string(bucketName))
			}

			cur := b.Cursor()
			for k, v := cur.First(); k != nil; k, v = cur.Next() {
				if v == nil {
					continue
				}

				oldItem, err := decodeOldCacheItem(v)
				if err != nil {
					log.Error(err)
				}

				if oldItem.ErrorCode == 0 && time.Since(time.Unix(oldItem.Timestamp, 0)) < getCacheExpiration(oldItem.ErrorCode) {
					item := Item{
						Text:        string(k),
						Translation: oldItem.Translation,
						ErrorCode:   oldItem.ErrorCode,
						ErrorText:   oldItem.ErrorText,
						Timestamp:   oldItem.Timestamp,
					}

					items = append(items, item)

					count++

					if count%10000 == 0 {
						log.Info(count)
					}
				}

				b.Delete(k)

			}

			log.Infof("read %d valid records", count)

			return nil
		})

		if count > 0 {
			count = 0

			db := c.db.From(string(bucketName))
			tx, err := db.Begin(true)
			if err != nil {
				return err
			}
			defer tx.Rollback()

			log.Infof("writing new records")

			for _, item := range items {

				err = tx.Save(&item)
				if err != nil {
					log.Fatal(err)
				}

				count++

				if count%10000 == 0 {
					log.Info(count)
				}
			}

			tx.Commit()

			log.Infof("%d records written", count)
		}

		if err != nil {
			log.Error("failed to migrate: ", string(bucketName))
			continue
		}

	}

	// Delete old metadata
	err := c.db.Bolt.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket([]byte("___metadata"))
		if err != nil {
			log.Error(err)
		}

		return nil
	})

	return err
}
