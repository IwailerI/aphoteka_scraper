package scraper

import (
	"aphoteka_scraper/manifest"
	"encoding/json"
	"errors"
	"log"

	"github.com/gocolly/colly"
)

func FetchData(input map[string]string) (manifest.Manifest, error) {
	var e []error

	c := colly.NewCollector(
		colly.AllowedDomains("www.apotheka.lv"),
	)

	const XPATH = `//script[@type="application/ld+json"]`

	available := make(map[string]manifest.Availability)

	c.OnXML(XPATH, func(x *colly.XMLElement) {
		var data []struct {
			Offers struct {
				Availability  string  `json:"availability"`
				Price         float64 `json:"price"`
				PriceCurrency string  `json:"priceCurrency"`
			} `json:"offers"`
		}
		err := json.Unmarshal([]byte(x.Text), &data)
		if err != nil {
			e = append(e, errors.Join(
				errors.New("when parsing url "+x.Request.URL.String()),
				err,
			))
			return
		}
		if len(data) == 0 {
			e = append(e, errors.Join(
				errors.New("when parsing url "+x.Request.URL.String()),
				errors.New("empty data"),
			))
			return
		}
		available[x.Request.URL.String()] = manifest.Availability{
			Price:    uint(data[0].Offers.Price * 100),
			Tag:      data[0].Offers.Availability,
			Currency: data[0].Offers.PriceCurrency,
		}

	})

	c.OnRequest(func(r *colly.Request) {
		log.Print("Visiting ", r.URL)
	})

	for _, url := range input {
		err := c.Visit(url)
		if err != nil {
			e = append(e, err)
		}
	}

	manifest := make(manifest.Manifest)
	for name, url := range input {
		v, ok := available[url]
		if !ok {
			log.Printf("Invalid url: %v", url)
		}

		v.Url = url

		manifest[name] = v
	}

	return manifest, errors.Join(e...)
}
