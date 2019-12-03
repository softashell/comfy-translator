package config

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Host     string
	Port     string
	Database struct {
		Engine string
		Sqlite struct {
			Path      string
			CacheSize int
		}
		PostgreSQL struct {
			URL string
		}
	}
	Translator map[string]TranslatorConfig
}

type TranslatorConfig struct {
	Enabled  bool
	Priority int
	Key      string
}

func NewConfig() *Config {
	conf := createDefaultConfig()

	return &conf
}

func (c *Config) Load(configPath string) error {
	var nc Config

	var _, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		c.Save(configPath)
		return nil
	}

	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Error(err)
		return err
	}

	tomlText := string(dat)
	if _, err := toml.Decode(tomlText, &nc); err != nil {
		log.Error(err)
		return err
	}

	if len(nc.Host) > 0 {
		c.Host = nc.Host
	}

	if len(nc.Port) > 0 {
		c.Port = nc.Port
	}

	if len(nc.Database.Engine) > 0 {
		c.Database.Engine = nc.Database.Engine
	}

	if len(nc.Database.Sqlite.Path) > 0 {
		c.Database.Sqlite.Path = nc.Database.Sqlite.Path
	}

	if len(nc.Database.Sqlite.Path) > 0 {
		c.Database.Sqlite.CacheSize = nc.Database.Sqlite.CacheSize

		if c.Database.Sqlite.CacheSize < 2000 {
			c.Database.Sqlite.CacheSize = 2000
		}
	}

	if len(nc.Database.PostgreSQL.URL) > 0 {
		c.Database.PostgreSQL.URL = nc.Database.PostgreSQL.URL
	}

	for k, v := range nc.Translator {
		c.Translator[k] = v
	}

	return nil
}

func (c *Config) Save(configPath string) {
	f, err := os.Create(configPath)
	if err != nil {
		log.Error("Failed to save default config ", err)
	}
	defer f.Close()

	e := toml.NewEncoder(f)
	e.Encode(c)
}

func createDefaultConfig() Config {
	var c Config

	c.Host = "127.0.0.1"
	c.Port = "3000"

	c.Database.Engine = "sqlite"

	c.Database.Sqlite.Path = "translation.db"
	c.Database.Sqlite.CacheSize = 125000
	c.Database.PostgreSQL.URL = "postgres://username:password@127.0.0.1:5432/database"

	t := make(map[string]TranslatorConfig)

	t["Google"] = TranslatorConfig{
		Enabled:  true,
		Priority: 1,
	}

	t["Bing"] = TranslatorConfig{
		Enabled:  false,
		Priority: 2,
	}

	t["Yandex"] = TranslatorConfig{
		Enabled:  false,
		Priority: 3,
	}

	c.Translator = t

	return c
}
