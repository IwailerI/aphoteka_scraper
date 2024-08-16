package telegram

import (
	"aphoteka_scraper/permanence"
	"aphoteka_scraper/secrets"
	"bytes"
	"encoding/gob"
	"os"
	"path"
	"time"
)

type serverConfig struct {
	Whitelist       map[string]struct{}
	NotifyChannels  []string
	ServiceChannels []string
	Products        map[string]string
	Active          bool
	Interval        time.Duration
}

var config serverConfig
var unit = struct{}{}

func newServerConfig() serverConfig {
	return serverConfig{
		Whitelist:       map[string]struct{}{secrets.RootUser: unit},
		NotifyChannels:  []string{},
		ServiceChannels: []string{},
		Products:        map[string]string{},
		Active:          true,
		Interval:        1 * time.Hour,
	}
}

func loadServerConfig() error {
	p, err := permanence.GetUserDir()
	if err != nil {
		return err
	}
	p = path.Join(p, "config.gob")

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			config = newServerConfig()
			return nil
		} else {
			return err
		}
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	c := newServerConfig()
	err = dec.Decode(&c)
	if err != nil {
		return err
	}

	c.Whitelist[secrets.RootUser] = unit

	config = c

	return nil
}

func saveServerConfig() error {
	p, err := permanence.GetUserDir()
	if err != nil {
		return err
	}
	p = path.Join(p, "config.gob")

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	err = enc.Encode(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile(p, buf.Bytes(), 0666)
	if err != nil {
		return err
	}

	return nil
}
