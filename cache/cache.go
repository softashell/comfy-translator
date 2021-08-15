package cache

import (
	"fmt"
	"strings"
	"time"

	"gitgud.io/softashell/comfy-translator/cache/postgres"
	"gitgud.io/softashell/comfy-translator/cache/sqlite"
	"gitgud.io/softashell/comfy-translator/config"
)

type translationError int

const (
	ErrorNone           translationError = iota // Everything is fine
	ErrorMinor                                  // Connection timed out or something like that
	ErrorBadTranslation                         // Returned really bad translation
)

const (
	CleanupInterval = (24 * time.Hour) * 7
)

type Item struct {
	Translation string
	ErrorCode   translationError
	ErrorText   string
	Timestamp   int64
}

type Cache interface {
	Put(bucketName, text, translation string, cerr error) error
	Get(bucketName, text string) (string, bool, error)
	Close() error
}

func NewCache(conf *config.Config, translators []string) (Cache, error) {

	engineName := strings.ToLower(conf.Database.Engine)

	switch engineName {
	case "sqlite":
		return sqlite.NewCache(conf.Database.Sqlite.Path, conf.Database.Sqlite.CacheSize, translators)
	case "postgresql","postgres":
		return postgres.NewCache(conf.Database.PostgreSQL.URL, conf.Database.Sqlite.CacheSize, translators)
	}

	return nil, fmt.Errorf("unknown database engine %s", engineName)
}
