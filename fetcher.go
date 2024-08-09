package main

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/gocolly/colly"
)

func fetch_data(input map[string]string) (map[string]Availability, error) {

	var e []error

	c := colly.NewCollector(
		colly.AllowedDomains("www.apotheka.lv"),
	)

	const XPATH = `//script[@type="application/ld+json"]`

	available := make(map[string]Availability)

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
		available[x.Request.URL.String()] = Availability{
			price:    uint(data[0].Offers.Price * 100),
			tag:      data[0].Offers.Availability,
			currency: data[0].Offers.PriceCurrency,
		}
		log.Println(available)

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

	output := make(map[string]Availability)
	for name, url := range input {
		v, ok := available[url]
		if !ok {
			log.Printf("Invalid url: %v", url)
		}

		v.url = url

		output[name] = v
	}

	return output, errors.Join(e...)
}
