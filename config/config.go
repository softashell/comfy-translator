package config

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

type Config struct {
	Host     string
	Port     string
	Database struct {
		Path string
	}
	Translator map[string]TranslatorConfig
}

type TranslatorConfig struct {
	Enabled  bool
	Priority int
	Key      string
}

func NewConfig() *Config {
	var conf Config

	conf = createDefaultConfig()

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

	if len(nc.Database.Path) > 0 {
		c.Database.Path = nc.Database.Path
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

	c.Database.Path = "translation.db"

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
