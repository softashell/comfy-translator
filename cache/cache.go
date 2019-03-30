package cache

import (
	"database/sql"
	"fmt"
	"time"

	"gitgud.io/softashell/comfy-translator/translator"
	lru "github.com/hashicorp/golang-lru"
	_ "github.com/mattn/go-sqlite3" // Sql driver
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
	db       *sql.DB
	lrustore map[string]*lru.TwoQueueCache
}

type Item struct {
	Translation string
	ErrorCode   translationError
	ErrorText   string
	Timestamp   int64
}

func NewCache(filePath string, cacheSize int, translators []string) (*Cache, error) {
	db, err := sql.Open("sqlite3", filePath+"?cache=shared&mode=rwc&_synchronous=1&_auto_vacuum=2&_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Avoids database is locked errors
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(fmt.Sprintf("PRAGMA cache_size = -%d", cacheSize)); err != nil {
		log.Errorf("Failed to increase page cache size! %s", err)
	} else {
		log.Infof("Page cache set to %d kib", cacheSize)
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

	cache.migrateDatabase()

	return cache, nil
}

func (c *Cache) Close() error {
	log.Println("Writing checkpoint")
	c.db.Exec("PRAGMA wal_checkpoint")

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

	stmt, err := c.db.Prepare("INSERT OR REPLACE INTO Translations(text, service, translation, errorCode, errorText, time) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(text, bucketName, translation, errorCode, errorText, time.Now().UTC().Unix())
	if err != nil {
		log.Fatal(err)
	}

	// Add to memory cache
	i := Item{
		Translation: translation,
		ErrorCode:   errorCode,
		ErrorText:   errorText,
		Timestamp:   time.Now().UTC().Unix(),
	}

	c.lrustore[bucketName].Add(text, i)

	return err
}

func (c *Cache) Get(bucketName, text string) (string, bool, error) {
	var found bool
	var id int64
	var translation string
	var errorCode translationError
	var errorText string
	var timestamp int64
	var err error

	item, ok := c.lrustore[bucketName].Get(text)
	if ok {
		i := item.(Item)

		translation = i.Translation
		errorCode = i.ErrorCode
		errorText = i.ErrorText
	} else {
		stmt, err := c.db.Prepare("SELECT id, translation, errorCode, errorText, time FROM Translations WHERE service = ? AND text = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		err = stmt.QueryRow(bucketName, text).Scan(&id, &translation, &errorCode, &errorText, &timestamp)
		if err != nil {
			return translation, found, nil
		}
	}

	if translation != "" {
		found = true
	}

	if errorCode != errorNone {
		errorTime := time.Unix(timestamp, 0)

		if time.Since(errorTime) > getCacheExpiration(errorCode) {
			if id != 0 {
				_, err = c.db.Exec(fmt.Sprintf("DELETE FROM Translations WHERE id = %d", id))
				if err != nil {
					log.Warn("unable to delete item: ", err)
				}
			}

			// Act as if nothing was found
			return "", false, nil
		}

		return "", found, fmt.Errorf("%s", errorText)
	}

	return translation, found, nil
}
