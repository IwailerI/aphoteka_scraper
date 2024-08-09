# What is this
This is a utility executable, that checks whether certain aphoteke products are available and notifies via telegram if they are.

# Setup
Create 2 files in the root of the repository: 
- `token.secret`: should contain discord bot api key
- `chats.secret`: should contain discord channel ids, one on each line

Each channel in `chats.secret` will be notified by the bot, whose token is provided in `token.secret`.

Currently 2 dexcom products are hard-coded into the main function.
Program creates a last_run.txt dump on each run. (It should be very small, < 1kB)

# Options
By default, bot will notify channels only when products are available. This can be overriden if `--force-notify` is provided as first command line argument.
