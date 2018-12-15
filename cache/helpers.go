package cache

import (
	"bytes"
	"encoding/json"
	"time"

	log "github.com/Sirupsen/logrus"
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

func decodeOldCacheItem(val []byte) (oldCacheItem, error) {
	var i oldCacheItem

	buf := bytes.NewReader(val)
	if err := json.NewDecoder(buf).Decode(&i); err != nil {
		// try to ignore decoding error and just log it
		log.Warnf("Failed to decode bytes into cacheItem; '%s' => %v", string(val), err)

		return i, err
	}

	return i, nil
}
