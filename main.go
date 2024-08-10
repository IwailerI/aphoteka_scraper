package main

import (
	"aphoteka_scraper/manifest"
	"aphoteka_scraper/permanence"
	"aphoteka_scraper/scraper"
	"aphoteka_scraper/telegram"
	"log"
	"os"
)

// Don't forget to update crontab!
var urls = map[string]string{
	"dexcom-sensor":        "https://www.apotheka.lv/dexcom-one-sensors-pmm0171563lv",
	"dexcom-sensor-3-pack": "https://www.apotheka.lv/dexcom-one-sensors-3pack-pmm0172453lv",
}

func main() {
	defer permanence.Logger.Save()

	work()

	if permanence.Logger.HasErrors() {
		err := telegram.SendErrors()
		if err != nil {
			log.Panicf("Cannot send errors via telegram: %v", err)
		}
	}
}

func work() {
	// parse command line options
	force_notify := len(os.Args) > 1 && os.Args[1] == "--force-notify"
	permanence.Logger.AddForceNotify(force_notify)

	// fetch the date, generate a new manifest
	m, err := scraper.FetchData(urls)
	if err != nil {
		permanence.Logger.AddError(err)
		return
	}
	permanence.Logger.AddManifest(m)

	// load previous manifest from disk
	prev_manifest, err := permanence.LoadManifest()
	if err != nil {
		log.Printf("Cannot access previous manifest: %v", err)
		permanence.Logger.AddError(err)
	}

	// check whether manifests match
	needs_update := !manifest.AreEqual(prev_manifest, m)

	// save new manifest to disk
	if needs_update {
		err = permanence.SaveManifest(m)
		if err != nil {
			permanence.Logger.AddError(err)
			log.Printf("Cannot save new manifest: %v", err)
		}
	}

	// generate a message to send, also print it to stdout
	msg := m.GenerateMessage()
	permanence.Logger.AddMessage(msg)
	log.Printf("Message:\n%v\n\n(end)", msg)

	// notify via telegram
	if needs_update || force_notify {
		err := telegram.SendUpdate(msg)
		if err != nil {
			permanence.Logger.AddError(err)
			log.Printf("Cannot send telegram update:\n%v", err)
		}
	}
}
