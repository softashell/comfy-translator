package postgres

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func getCacheExpiration(errorCode translationError) time.Duration {
	switch errorCode {
	case errorNone:
		return (24 * time.Hour) * 365
	case errorMinor:
		return time.Hour / 2
	case errorBadTranslation:
		return 24 * time.Hour
	}

	log.Fatalf("unknow error code: %d", errorCode)

	return time.Second
}
