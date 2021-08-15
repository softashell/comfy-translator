package postgres

import (
	"fmt"
	"time"

	"gitgud.io/softashell/comfy-translator/translator"
	lru "github.com/hashicorp/golang-lru"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type translationError int

const (
	errorNone           translationError = iota // Everything is fine
	errorMinor                                  // Connection timed out or something like that
	errorBadTranslation                         // Returned really bad translation
)

type Cache struct {
	db       *gorm.DB
	lrustore map[string]*lru.TwoQueueCache
}

type Item struct {
	Translation string
	ErrorCode   translationError
	ErrorText   string
	Timestamp   int64
}

type PgItem struct {
	Text string `gorm:"primarykey"`
	Bucket string `gorm:"primaryKey"`
	Translation string
	ErrorCode   translationError
	ErrorText   string
	Timestamp   time.Time
}

func NewCache(connStr string, cacheSize int, translators []string) (*Cache, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = db.AutoMigrate(&PgItem{})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	store := make(map[string]*lru.TwoQueueCache)
	for _, t := range translators {
		s, err := lru.New2Q(5000)
		if err != nil {
			log.Fatalf("Can't start memory cache for %s", t)
		}

		store[t] = s
	}

	cache := &Cache{db: db, lrustore: store}

	return cache, nil
}

func (c *Cache) Close() error {
	return nil
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

	pgItem := PgItem{
		Text: 			text,	
		Bucket: 		bucketName,
		Translation: 	translation,
		ErrorCode:   	errorCode,
		ErrorText:   	errorText,
		Timestamp:   	time.Now().UTC(),
	}

	result := c.db.Create(&pgItem)
	if result.Error != nil {
		log.Fatal(errors.Wrap(result.Error, "failed to execute insert"))
	}

	// Add to memory cache
	i := Item{
		Translation: translation,
		ErrorCode:   errorCode,
		ErrorText:   errorText,
		Timestamp:   time.Now().UTC().Unix(),
	}

	c.lrustore[bucketName].Add(text, i)

	return result.Error
}

func (c *Cache) Get(bucketName, text string) (string, bool, error) {
	var found bool
	var translation string
	var errorCode translationError
	var errorText string
	var timestamp int64

	item, ok := c.lrustore[bucketName].Get(text)
	if ok {
		i := item.(Item)

		translation = i.Translation
		errorCode = i.ErrorCode
		errorText = i.ErrorText
	} else {
		i := PgItem{}
		result := c.db.Limit(1).Find(&i, PgItem{Text: text,	Bucket: bucketName})
		if result.Error != nil {
			return translation, found, nil
		}

		translation = i.Translation
		errorCode = i.ErrorCode
		errorText = i.ErrorText
	}

	if translation != "" {
		found = true
	}

	if errorCode != errorNone {
		errorTime := time.Unix(timestamp, 0)

		if time.Since(errorTime) > getCacheExpiration(errorCode) {
			i := PgItem{}
			result := c.db.Delete(&i, PgItem{Text: text,	Bucket: bucketName})
			if result.Error != nil {
				log.Warn("unable to delete item: ", result.Error)
			}

			c.lrustore[bucketName].Remove(text)

			// Act as if nothing was found
			return "", false, nil
		}

		return "", found, fmt.Errorf("%s", errorText)
	}

	return translation, found, nil
}
