package telegram

import (
	"aphoteka_scraper/permanence"
	"aphoteka_scraper/scraper"
	"aphoteka_scraper/secrets"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var ErrorCannotSetCommands = errors.New("settings telegram bot commands failed")

var loopStopHandle chan<- struct{}
var loopStopHandleValid = false
var nextCheck time.Time
var lastCheck time.Time

func RunServer() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := loadServerConfig()
	if err != nil {
		return err
	}

	b, err := bot.New(secrets.Token)
	if err != nil {
		return err
	}

	err = setupCommands(ctx, b)
	if err != nil {
		return err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/status", bot.MatchTypePrefix, handleStatus)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/add_user", bot.MatchTypePrefix, handlerAddUser)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/remove_user", bot.MatchTypePrefix, handleRemoveUser)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/list_users", bot.MatchTypePrefix, handleListUsers)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/add_channel", bot.MatchTypePrefix, handleAddChannel)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/remove_channel", bot.MatchTypePrefix, handleRemoveChannel)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/add_service_channel", bot.MatchTypePrefix, handleAddServiceChannel)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/remove_service_channel", bot.MatchTypePrefix, handleRemoveServiceChannel)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/list_channels", bot.MatchTypePrefix, handleListChannels)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/add_product", bot.MatchTypePrefix, handleAddProduct)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/remove_product", bot.MatchTypePrefix, handleRemoveProduct)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/list_products", bot.MatchTypePrefix, handleListProducts)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/force_update", bot.MatchTypePrefix, handleForceUpdate)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/check_now", bot.MatchTypePrefix, handleCheckNow)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start_updates", bot.MatchTypePrefix, handleStartUpdates)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/stop_updates", bot.MatchTypePrefix, handleStopUpdates)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/set_update_interval", bot.MatchTypePrefix, handleSetUpdateInterval)

	if config.Active {
		setupLoop(ctx, b)
	}

	log.Print("Server started")

	b.Start(ctx)

	return nil
}

func setupCommands(ctx context.Context, b *bot.Bot) error {
	ok, err := b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "/status", Description: "Get status of the bot"},

			{Command: "/add_user", Description: "Add user to the whitelist"},
			{Command: "/remove_user", Description: "Remove user from whitelist"},
			{Command: "/list_users", Description: "List users currently in whitelist"},

			{Command: "/add_channel", Description: "Add channel to be notified"},
			{Command: "/remove_channel", Description: "Stops notifying a channel"},
			{Command: "/add_service_channel", Description: "Add channel for service notifications"},
			{Command: "/remove_service_channel", Description: "Stop notifying a channel"},
			{Command: "/list_channels", Description: "Lists all currently notified channels"},

			{Command: "/add_product", Description: "Adds a new product to be tracked"},
			{Command: "/remove_product", Description: "Stops tracking some product"},
			{Command: "/list_products", Description: "List currently tracked products"},

			{Command: "/force_update", Description: "Notify all channels, regardless of result"},
			{Command: "/check_now", Description: "Check for result, as if it was scheduled"},
			{Command: "/start_updates", Description: "Turns notifications and updates on"},
			{Command: "/stop_updates", Description: "Turns notifications and updates off"},
			{Command: "/set_update_interval", Description: "Sets update interval in minutes"},
		},
	})

	if err != nil {
		return err
	}

	if !ok {
		return ErrorCannotSetCommands
	}

	return nil
}

func checkPermission(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	if _, ok := config.Whitelist["@"+update.Message.From.Username]; ok {
		return true
	} else {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Unauthorized. Sorry.",
		})

		return false
	}
}

func swapRemove[T any](s []T, i int) []T {
	if len(s) == 1 {
		return nil
	}
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func setupLoop(ctx context.Context, b *bot.Bot) {
	log.Print("Setting up new loop...")

	if loopStopHandleValid {
		loopStopHandle <- unit
		close(loopStopHandle)
	}
	config.Active = true

	stop := make(chan struct{})
	loopStopHandle = stop
	loopStopHandleValid = true
	ticker := time.Tick(config.Interval)
	d := config.Interval
	nextCheck = time.Now().Add(d)

	go func() {
		for {
			select {
			case now := <-ticker:
				nextCheck = now.Add(d)
				checkAndNotify(ctx, b, false)
			case <-stop:
				break
			}
		}
	}()
}

func checkAndNotify(ctx context.Context, b *bot.Bot, forceUpdate bool) {
	newManifest, needsUpdate, err := scraper.FetchAndCompare(config.Products)
	lastCheck = time.Now()
	error_slice := []error{}

	msg := newManifest.GenerateMessage()

	log.Print(msg)
	if err != nil {
		error_slice = append(error_slice, err)
	}

	if needsUpdate || forceUpdate {
		for _, channel := range config.NotifyChannels {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: channel,
				Text:   msg,
			})
			if err != nil {
				error_slice = append(error_slice, err)
			}
		}
	}

	if len(error_slice) > 0 {
		m := fmt.Sprintf("[SERVICE]\n%v", errors.Join(error_slice...))
		log.Print(m)

		for _, channel := range config.ServiceChannels {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: channel,
				Text:   m,
			})
			if err != nil {
				error_slice = append(error_slice, err)
			}
		}
	}
}

func handlerAddUser(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	username, ok := strings.CutPrefix(update.Message.Text, "/add_user ")
	username = "@" + strings.TrimPrefix(strings.TrimSpace(username), "@")

	if !ok || len(username) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_user <username>",
		})
		return
	}

	config.Whitelist[username] = unit
	err := saveServerConfig()
	if err != nil {
		log.Printf("Cannot save server config! %v", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("User %q added.", username),
	})
}

func handleRemoveUser(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	username, ok := strings.CutPrefix(update.Message.Text, "/remove_user ")
	username = "@" + strings.TrimPrefix(strings.TrimSpace(username), "@")

	if !ok || len(username) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_user <username>",
		})
		return
	}

	if username == secrets.RootUser {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Cannot remove root user. (yay, you found who is root)",
		})
		return
	}

	_, found := config.Whitelist[username]
	if !found {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("User %q is not in whitelist.", username),
		})
		return
	}

	delete(config.Whitelist, username)
	err := saveServerConfig()
	if err != nil {
		log.Printf("Cannot save server config! %v", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("User %q removed.", username),
	})
}

func handleListUsers(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	keys := []string{}
	for user := range config.Whitelist {
		keys = append(keys, user)
	}
	sort.Strings(keys)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Whitelist: %q", keys),
	})
}

func handleAddChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/add_channel ")
	channel = strings.TrimSpace(channel)

	if !ok || len(channel) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_channel <channel>",
		})
		return
	}

	if !slices.Contains(config.NotifyChannels, channel) {
		config.NotifyChannels = append(config.NotifyChannels, channel)
	}
	err := saveServerConfig()
	if err != nil {
		log.Printf("Cannot save server config! %v", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Channel %q added.", channel),
	})
}

func handleRemoveChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/remove_channel")
	channel = strings.TrimSpace(channel)

	if !ok || len(channel) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_channel <channel>",
		})
		return
	}

	i := slices.Index(config.NotifyChannels, channel)
	if i == -1 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Channel %q is not in found.", channel),
		})
		return
	}

	config.NotifyChannels = swapRemove(config.NotifyChannels, i)
	err := saveServerConfig()
	if err != nil {
		log.Printf("Cannot save server config! %v", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Channel %q will not be notified anymore.", channel),
	})
}

func handleAddServiceChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/add_service_channel ")

	if !ok || len(channel) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_service_channel <channel>",
		})
		return
	}

	if !slices.Contains(config.ServiceChannels, channel) {
		config.ServiceChannels = append(config.ServiceChannels, channel)
		log.Print(config.ServiceChannels)
	}
	err := saveServerConfig()
	if err != nil {
		log.Printf("Cannot save server config! %v", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Service channel %q added.", channel),
	})
}

func handleRemoveServiceChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/remove_service_channel ")

	if !ok || len(channel) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_service_channel <channel>",
		})
		return
	}

	i := slices.Index(config.ServiceChannels, channel)
	if i == -1 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Channel %q is not in found.", channel),
		})
		return
	}

	config.ServiceChannels = swapRemove(config.ServiceChannels, i)
	err := saveServerConfig()
	if err != nil {
		log.Printf("Cannot save server config! %v", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Service channel %q will not be notified anymore.", channel),
	})
}

func handleListChannels(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	sort.Strings(config.NotifyChannels)
	sort.Strings(config.ServiceChannels)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Notify: %q\nService: %q", config.NotifyChannels, config.ServiceChannels),
	})
}

func handleForceUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Updating",
	})

	checkAndNotify(ctx, b, true)

}

func handleStartUpdates(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	if !config.Active {
		setupLoop(ctx, b)
		config.Active = true
		err := saveServerConfig()
		if err != nil {
			log.Printf("Cannot save server config! %v", err)
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Update cycle started.",
	})
}

func handleStopUpdates(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	if config.Active {
		loopStopHandle <- unit
		config.Active = false
		err := saveServerConfig()
		if err != nil {
			log.Printf("Cannot save server config! %v", err)
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Update cycle stopped.",
	})
}

func handleSetUpdateInterval(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	s, ok := strings.CutPrefix(update.Message.Text, "/set_update_interval ")
	if !ok {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /set_update_interval <number of minutes>",
		})
		return
	}
	s = strings.TrimSpace(s)

	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Expected positive integer as first argument",
		})
		return
	}

	config.Interval = time.Duration(n) * time.Minute
	err = saveServerConfig()
	if err != nil {
		log.Printf("Cannot save server config! %v", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Interval updated to %d hours %d minutes.", config.Interval/time.Hour, config.Interval%time.Hour/time.Minute),
	})
}

func handleAddProduct(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	s, ok := strings.CutPrefix(update.Message.Text, "/add_product ")
	slice := strings.Fields(s)
	if !ok || len(slice) != 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_product <name_of_product> <url>",
		})
		return
	}

	prev_url, overridden := config.Products[slice[0]]

	if !(overridden && prev_url == slice[1]) {
		config.Products[slice[0]] = slice[1]

		err := saveServerConfig()
		if err != nil {
			log.Printf("Cannot save server config! %v", err)
		}
	}

	if overridden {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: fmt.Sprintf("Product %q now has url %q instead of %q.",
				slice[0],
				slice[1],
				prev_url,
			),
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: fmt.Sprintf("Product %q added with url %q.",
				slice[0],
				slice[1],
			),
		})
	}
}

func handleRemoveProduct(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	s, ok := strings.CutPrefix(update.Message.Text, "/remove_product ")
	s = strings.TrimSpace(s)
	if !ok || s == "" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_product <name_of_product>",
		})
		return
	}

	_, found := config.Products[s]
	if !found {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Product %q is not found.", s),
		})
		return
	}
	delete(config.Products, s)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Product %q is deleted.", s),
	})
}

func handleListProducts(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	var s strings.Builder

	for product, url := range config.Products {
		fmt.Fprintf(&s, "%s - %s\n", product, url)
	}

	if s.Len() == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "There are no products tracked.",
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   s.String(),
		})
	}
}

func handleStatus(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	var s strings.Builder

	configDump, err := json.MarshalIndent(config, "", "    ")
	if err == nil {
		fmt.Fprintf(&s, "Current config:\n```json\n%s\n```\n", configDump)
	} else {
		log.Printf("Cannot create config dump: %v", err)
	}

	fmt.Fprintf(&s, "Human readable interval: %v\n", config.Interval)

	lastManifest, err := permanence.LoadManifest()
	if err == nil {
		fmt.Fprintf(&s, "Last manifest:\n%s\n\n", bot.EscapeMarkdown(lastManifest.GenerateMessage()))
	} else {
		log.Printf("Cannot open previous manifest: %v", err)
	}

	if !lastCheck.IsZero() {
		fmt.Fprintf(&s, "Last check:\n`%v`\n",
			bot.EscapeMarkdown(fmt.Sprint(lastCheck)))
	}

	if !nextCheck.IsZero() {
		fmt.Fprintf(&s, "Next check scheduled for:\n`%v`\n",
			bot.EscapeMarkdown(fmt.Sprint(nextCheck)))
	}

	fmt.Fprintf(&s,
		"Channels: %d\nService channels: %d\nProducts: %d\nAdmins: %d\n",
		len(config.NotifyChannels),
		len(config.ServiceChannels),
		len(config.Products),
		len(config.Whitelist),
	)

	log.Println(s.String())

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      s.String(),
		ParseMode: models.ParseModeMarkdown,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
	})
}

func handleCheckNow(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Updating",
	})

	checkAndNotify(ctx, b, false)
}
