package cache

import (
	"time"

	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
)

type StormItem struct {
	Text        string `storm:"id"` // primary key
	Translation string
	ErrorCode   translationError `storm:"index"`
	ErrorText   string
	Timestamp   int64
}

const latestVersion = 1

func (c *Cache) migrateDatabase() error {
	latestMigration := 0

	rows, err := c.db.Query("SELECT * FROM Migrations")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&latestMigration); err != nil {
				log.Fatal(err)
			}

			log.Printf("Migration #%d already in place", latestMigration)
		}
	}

	if latestVersion-latestMigration == 0 {
		return nil
	}

	log.Infof("Need to run %d database migrations", latestVersion-latestMigration)

	for latestMigration < latestVersion {
		latestMigration = c.migrate(latestMigration)
	}

	return nil
}

func (c *Cache) migrate(ver int) int {
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

	log.Info("Finished migration")

	return tgt
}

func (c *Cache) migration1() error {
	log.Print("Migration #1")

	tx, err := c.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	if err = execTxAndPrint(tx,
		`CREATE TABLE IF NOT EXISTS Migrations (
		id INT PRIMARY KEY NOT NULL
		);`); err != nil {
		return err
	}

	if err = execTxAndPrint(tx,
		`CREATE TABLE IF NOT EXISTS Translations (
			id INTEGER PRIMARY KEY,
			text TEXT NOT NULL,
			service TEXT NOT NULL,
			translation TEXT NOT NULL,
			errorCode INT,
			errorText TEXT,
			time INT
			);`); err != nil {
		return err
	}

	if err = execTxAndPrint(tx,
		`CREATE UNIQUE INDEX "translation_idx" ON "Translations" (
			"text",
			"service"
			);`); err != nil {
		return err
	}

	if err = execTxAndPrint(tx, `INSERT INTO migrations VALUES (1)`); err != nil {
		return err
	}

	err = tx.Commit()

	c.migrateFromStorm()

	return err
}

func (c *Cache) migrateFromStorm() {
	storm, err := storm.Open("_translation.db", storm.Batch())
	if err != nil {
		log.Error("can't open legacy storm db (_translation.db) for import")
		return
	}

	buckets := []string{
		"Bing",
		"Yandex",
		"Google",
	}

	for _, bucketName := range buckets {
		var items []StormItem

		db := storm.From(bucketName)
		err := db.All(&items)
		if err != nil {
			log.Error(err)
		}

		size := len(items)

		tx, err := c.db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		stmt, err := tx.Prepare("INSERT INTO Translations(text, service, translation, errorCode, errorText, time) VALUES (?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		for i, item := range items {
			log.Printf("%s: %d / %d", bucketName, i, size)

			_, err = stmt.Exec(item.Text, bucketName, item.Translation, item.ErrorCode, item.ErrorText, time.Now().UTC().Unix())
			if err != nil {
				log.Warnln(err)
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}
}
