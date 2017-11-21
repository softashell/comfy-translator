package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"gitgud.io/softashell/comfy-translator/translator"

	"bytes"

	log "github.com/Sirupsen/logrus"
	bolt "github.com/coreos/bbolt"
)

type translationError int

const (
	errorNone           translationError = iota // Everything is fine
	errorMinor                                  // Connection timed out or something like that
	errorBadTranslation                         // Returned really bad translation
)

const (
	cleanupInterval = (24 * time.Hour) * 7
)

type Cache struct {
	db *bolt.DB

	meta cacheMetadata
}

type cacheItem struct {
	Translation string           `json:"t"`
	ErrorCode   translationError `json:"c,omitempty"`
	ErrorText   string           `json:"e,omitempty"`
	Timestamp   int64            `json:"u,omitempty"`
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

	go cache.cleanCacheEntries()

	return cache, nil
}

func (c *Cache) Close() error {
	return c.db.Close()
}

func (c *Cache) Put(bucketName, text, translation string, cerr error) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket doesn't exist??")
		}

		errorCode := errorNone
		errorText := ""

		if cerr != nil {
			errorText = cerr.Error()

			switch cerr.(type) {
			case translator.BadTranslationError:
				errorCode = errorBadTranslation
			default:
				errorCode = errorMinor
			}
		}

		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(cacheItem{
			Translation: translation,
			ErrorCode:   errorCode,
			ErrorText:   errorText,
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
	var err error
	var found bool

	err = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			log.Warnf("Bucket %q not found! (Probably nothing has been saved to it yet)", bucketName)
			return nil
		}

		val := b.Get([]byte(text))
		if val == nil {
			return nil
		}

		i, err = decodeCacheItem(val)
		if err == nil {
			found = true
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
		return "", false, err
	}

	if !found {
		return "", false, nil
	}

	if i.ErrorCode != errorNone {
		errorTime := time.Unix(i.Timestamp, 0)

		if time.Since(errorTime) > getCacheExpiration(i.ErrorCode) {
			return "", false, nil
		}

		return "", true, fmt.Errorf("%s", i.ErrorText)
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

func (c *Cache) cleanCacheEntries() error {
	if time.Since(time.Unix(c.meta.LastCleanup, 0)) < cleanupInterval {
		log.Info("Skipping cache cleanup")
		return nil
	}

	log.Info("Cleaning up old cache entries")

	buckets, err := c.getBuckets()
	if err != nil {
		return err
	}

	removalList := make(map[string][][]byte)

	err = c.db.View(func(tx *bolt.Tx) error {
		for _, bucketName := range buckets {
			b := tx.Bucket(bucketName)
			if b == nil {
				return nil
			}

			c := b.Cursor()

			for key, value := c.First(); key != nil; key, value = c.Next() {
				if value == nil {
					continue
				}

				item, err := decodeCacheItem(value)
				if err != nil {
					removalList[string(bucketName)] = append(removalList[string(bucketName)], key)
				}

				if item.ErrorCode != errorNone || item.ErrorText != "" {
					errorTime := time.Unix(item.Timestamp, 0)

					if item.ErrorCode == errorNone {
						item.ErrorCode = errorMinor
					}

					if time.Since(errorTime) > getCacheExpiration(item.ErrorCode) {
						removalList[string(bucketName)] = append(removalList[string(bucketName)], key)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Error(err)
		return err
	}

	if len(removalList) < 1 {
		log.Info("Nothing to remove")
		return nil
	}

	err = c.db.Update(func(tx *bolt.Tx) error {
		for _, bucketName := range buckets {
			b := tx.Bucket(bucketName)
			if b == nil {
				return nil
			}

			removed := 0

			for _, k := range removalList[string(bucketName)] {
				if err := b.Delete(k); err != nil {
					log.Warnf("failed to delete cache entry: %v", err)
				}

				removed++
			}

			if removed > 0 {
				log.Infof("Removed %d out of %d expired or invalid cache entries from %s", removed, len(removalList[string(bucketName)]), string(bucketName))
			}
		}

		return nil
	})

	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("Cache cleanup done")

	c.meta.LastCleanup = time.Now().UTC().Unix()

	if err := c.writeMetadata(); err != nil {
		log.Error(err)
	}

	return nil
}
