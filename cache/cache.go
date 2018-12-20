package cache

import (
	"fmt"
	"time"

	"gitgud.io/softashell/comfy-translator/translator"

	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
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
	db *storm.DB

	meta Metadata
}

type Item struct {
	Text        string `storm:"id"` // primary key
	Translation string
	ErrorCode   translationError `storm:"index"`
	ErrorText   string
	Timestamp   int64
}

func NewCache(filePath string) (*Cache, error) {
	db, err := storm.Open(filePath, storm.Batch())
	//db, err := bolt.Open(filePath, 0600, nil)
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

func (c *Cache) Put(bucketName, text, translation string, cerr error) error {
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

	item := Item{
		Text:        text,
		Translation: translation,
		ErrorCode:   errorCode,
		ErrorText:   errorText,
		Timestamp:   time.Now().UTC().Unix(),
	}

	db := c.db.From(bucketName)
	err := db.Save(&item)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (c *Cache) Get(bucketName, text string) (string, bool, error) {
	var item Item
	var found bool

	db := c.db.From(bucketName)
	err := db.One("Text", text, &item)
	if err != nil {
		return "", false, err
	}

	if item.Translation != "" {
		found = true
	}

	if item.ErrorCode != errorNone {
		errorTime := time.Unix(item.Timestamp, 0)

		if time.Since(errorTime) > getCacheExpiration(item.ErrorCode) {
			err = db.DeleteStruct(&item)
			if err != nil {
				log.Warn("unable to delete item:", err)
			}

			// Act as if nothing was found
			return "", false, nil
		}

		return "", found, fmt.Errorf("%s", item.ErrorText)
	}

	return item.Translation, found, nil
}
