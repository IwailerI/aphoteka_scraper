package main

import (
	"aphoteka_scraper/telegram"
	"log"
)

func main() {
	err := telegram.RunServer()
	if err != nil {
		log.Print(err)
	}
	log.Print("Server shut down")
}
