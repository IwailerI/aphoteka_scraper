package permanence

import (
	"aphoteka_scraper/manifest"
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

var Logger = NewLog()

func getManifestFilename() (string, error) {
	filename, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(filename, "aphoteka_scraper/last_manifest.gob"), nil
}

func SaveManifest(data manifest.Manifest) error {
	filename, err := getManifestFilename()
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(filename), 0750)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := gob.NewEncoder(f)

	err = w.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

// Load tries to load the saved manifest. If the manifest does not exist,
// returns an empty manifest.
func LoadManifest() (manifest.Manifest, error) {
	filename, err := getManifestFilename()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return manifest.Manifest{}, nil
		} else {
			return nil, err
		}
	}
	defer f.Close()

	r := gob.NewDecoder(f)

	var data manifest.Manifest
	err = r.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type Log struct {
	forceNotify      *bool
	manifest         *manifest.Manifest
	message          *string
	telegramNotified bool
	err              []error
}

func NewLog() Log {
	return Log{}
}

func (l *Log) AddForceNotify(v bool) {
	l.forceNotify = &v
}

func (l *Log) AddManifest(m manifest.Manifest) {
	l.manifest = &m
}

func (l *Log) AddMessage(m string) {
	l.message = &m
}

func (l *Log) TelegramNotified() {
	l.telegramNotified = true
}

func (l *Log) AddError(err error) {
	l.err = append(l.err, err)
}

func (l *Log) HasErrors() bool {
	return len(l.err) > 0
}

func (l *Log) GetErrors() []error {
	return l.err
}

func (l *Log) Save() {
	filename, err := os.UserCacheDir()
	if err != nil {
		log.Printf("Cannot save log: %v", err)
		return
	}

	filename = path.Join(filename, "aphoteka_scraper/last_run.txt")

	log.Print("Saving log to " + filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Cannot save log: %v", err)
		return
	}
	defer f.Close()

	if l.message == nil && l.manifest != nil {
		l.AddMessage(l.manifest.GenerateMessage())
	}

	bf := bufio.NewWriter(f)
	defer bf.Flush()

	fmt.Fprintf(bf, "Timestamp: %v\n", time.Now())

	if l.forceNotify != nil {
		fmt.Fprintf(bf, "Force notify: %v\n", *l.forceNotify)
	}

	fmt.Fprintf(bf, "Telegram notified: %v\n", l.telegramNotified)

	if l.manifest != nil {
		fmt.Fprintf(bf, "Manifest dump: %v\n", *l.manifest)
	}

	if l.message != nil {
		fmt.Fprintf(bf, "Message:\n%v\n", *l.message)
	}

	if len(l.err) > 0 {
		fmt.Fprintln(bf, "Errors:")
		for i, err := range l.err {
			fmt.Fprintf(bf, "%d: %v\n", i, err)
		}
	}

}
