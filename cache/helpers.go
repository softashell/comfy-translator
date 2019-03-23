package cache

import (
	"database/sql"
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

func execAndPrint(db *sql.DB, stmt string) {
	log.Print(stmt)
	if _, err := db.Exec(stmt); err != nil {
		log.Fatalf("Failed! %s", err)
	}
}

func execTxAndPrint(tx *sql.Tx, stmt string) error {
	log.Print(stmt)

	if _, err := tx.Exec(stmt); err != nil {
		tx.Rollback()

		log.Error("Failed! %s", err)

		return err
	}

	return nil
}
