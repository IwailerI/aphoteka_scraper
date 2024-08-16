package scraper

import (
	"aphoteka_scraper/manifest"
	"aphoteka_scraper/permanence"
	"errors"
	"log"
)

func FetchAndCompare(urls map[string]string) (newManifest manifest.Manifest, needsUpdate bool, e error) {

	ers := []error{}

	// fetch the date, generate a new manifest
	m, err := FetchData(urls)
	if err != nil {
		e = err
		return
	}
	newManifest = m

	// load previous manifest from disk
	prev_manifest, err := permanence.LoadManifest()
	if err != nil {
		log.Printf("Cannot access previous manifest: %v", err)

		ers = append(ers, err)
	}

	// check whether manifests match
	needsUpdate = !manifest.AreEqual(prev_manifest, m)

	// save new manifest to disk
	if needsUpdate {
		err = permanence.SaveManifest(m)
		if err != nil {
			// permanence.Logger.AddError(err)
			ers = append(ers, err)
			log.Printf("Cannot save new manifest: %v", err)
		}
	}

	e = errors.Join(ers...)

	return
}
