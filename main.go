package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

type Availability struct {
	price    uint
	tag      string
	url      string
	currency string
}

func handle_error(err error) {
	new_err := send_update(fmt.Sprintf("Error: %s", err))
	if new_err != nil {
		log.Panicf("Cannot send telegram update:\n%v", err)
	}
}

// Don't forget to update crontab!

func main() {
	force_notify := len(os.Args) > 1 && os.Args[1] == "--force-notify"

	last_run, err := os.Create(path.Join(path.Dir(os.Args[0]), "last_run.txt"))
	if err != nil {
		log.Printf("Cannot create file: %v", err)
	} else {
		defer last_run.Close()
		fmt.Fprintf(last_run, "Last run: %v\nForce notify: %v\n",
			time.Now(), force_notify)
	}

	r, err := fetch_data(map[string]string{
		"dexcom-sensor":        "https://www.apotheka.lv/dexcom-one-sensors-pmm0171563lv",
		"dexcom-sensor-3-pack": "https://www.apotheka.lv/dexcom-one-sensors-3pack-pmm0172453lv",
	})
	if err != nil {
		handle_error(err)
		return
	}

	needs_update := false
	for _, av := range r {
		if av.tag != "https://schema.org/OutOfStock" {
			needs_update = true
			break
		}
	}

	msg := generate_message(r)
	log.Printf("Message:\n%v\n\n(end)", msg)

	if last_run != nil {
		fmt.Fprintf(last_run, "Fetch result: %v\nMessage:\n%v\n", r, msg)
	}

	if needs_update || force_notify {
		err := send_update(msg)
		if err != nil {
			log.Panicf("Cannot send telegram update:\n%v", err)
		}
	}

}
