package permanence

import (
	"aphoteka_scraper/manifest"
	"encoding/gob"
	"os"
	"path"
)

func GetUserDir() (string, error) {
	filename, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(filename, "aphoteka_scraper"), nil
}

func getManifestFilename() (string, error) {
	filename, err := GetUserDir()
	if err != nil {
		return "", err
	}

	return path.Join(filename, "last_manifest.gob"), nil
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
