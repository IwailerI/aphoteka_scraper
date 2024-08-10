# What is this
This is a utility executable, that checks whether certain aphoteka products are 
available and notifies via telegram if they are.
The program will notify on any change in product status or any error encountered
by it.

# Setup
Create 2 files in the root of the repository: 
- `secrets/token.secret`: should contain discord bot api key
- `sercrest/chats.secret`: should contain discord channel ids, one on each line

Each channel in `chats.secret` will be notified by the bot, whose token is 
provided in `token.secret`.

Currently 2 dexcom products are hard-coded into the main file.

# Implementation
Program creates 2 files on each run: both in `~/.config/aphoteka_scraper` on linux
and `%LocalAppData%/aphoteka_scraper` on windows.
`last_run.txt` contains some info about the last run in human readable format,
`manifest.gob` contains the last manifest serialized using [GOV](https://pkg.go.dev/encoding/gob).

- `package manifest` declares the manifest type.
- `package permanence` implements simple logging as well as manifest file IO.
- `package scraper` implements actual scraping from the aphoteka website.
- `package secrets` embeds sensitive data. I was too lazy to setup proper .env.
- `package telegram` implements message sending via telegram.

And everything is driven from `main.go`.


# Options
By default, bot will notify channels only when products are available. This can 
be overridden if `--force-notify` is provided as first command line argument.
