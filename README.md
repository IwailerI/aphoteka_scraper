# The problem
Certain e-comerse web site silently changes prices or delists products completly. I would like to track the changes of the price.

## The solution
Create a server which watches the prices, records the differences and notifies me.

This is a telegram bot, that checks whether certain aphoteka products are 
available and notifies via telegram if they are, or their prices changes.
The program will notify on any change in product status or any error encountered
by it.

# Setup
Create 2 files: 
- `secrets/token.secret`: should contain discord bot api key
- `sercrest/root_user.secret`: should contain singular telegram username: root
user, you will set up all of the settings from that account

# Configuration
Start messaging the bot. It will have a lot of commands. Here is a partial 
breakdown:

- add / remove / list users - only users in this list will be able to interact
with the bot. Root user is always in this list
- add / remove / list (possible service) channels - there are 2 types of channels:
notification and service. Notification channels only get product updates, service
channels only get error logs and so on.
- add / remove / list products: each product consists of a unique name and a url,
only added products will be tracked
- start / stop notifications: manage notifications or temporarily
disable them
- set interval: change how often aphoteka is queried
- check now: ignore interval and check now
- force update: ignore interval, check now and notify regardless of result

# Implementation
Bot has a 2 main files for permanens: `config.gob` and `manifest.gob`, both 
encoded using [GOB](https://pkg.go.dev/encoding/gob). They are located in 
`~/.config/aphoteka_scraper` on linux and `%LocalAppData%/aphoteka_scraper` on 
windows.
`config.gob` contains all settings that were configured.
`manifest.gob` contains the last manifest fetched.

- `package manifest` declares the manifest type.
- `package permanence` implements manifest file IO.
- `package scraper` implements actual scraping from the aphoteka website.
- `package secrets` embeds sensitive data. I was too lazy to setup proper .env.
- `package telegram` implements message sending via telegram and the interactive 
server.
