package telegram

import (
	"aphoteka_scraper/scraper"
	"aphoteka_scraper/secrets"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var ErrorCannotSetCommands = errors.New("settings telegram bot commands failed")
var ErrorCannotSend = errors.New("cannot send message")
var ErrorCannotSave = errors.New("cannot save server config")
var ErrorCannotDumpManifest = errors.New("cannot create manifest dump")
var ErrorCannotLoadManifest = errors.New("cannot open previous manifest file")

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
	notifyService(ctx, b, "Server started")

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
		log.Printf("Command: %q", update.Message.Text)
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
	ticker := time.NewTicker(config.Interval)
	d := config.Interval
	nextCheck = time.Now().Add(d)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case now := <-ticker.C:
				nextCheck = now.Add(d)
				checkAndNotify(ctx, b, false)
			case <-stop:
				return
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
				LinkPreviewOptions: &models.LinkPreviewOptions{
					IsDisabled: bot.True(),
				},
			})
			if err != nil {
				error_slice = append(error_slice, err)
			}
		}
	}

	handleError(ctx, b, errors.Join(error_slice...))
}

func handleError(ctx context.Context, b *bot.Bot, err error) {
	if err == nil {
		return
	}

	m := fmt.Sprintf("[SERVICE]\n%v", err)
	log.Print(m)

	err = notifyService(ctx, b, m)

	if err != nil {
		log.Printf("Error sending service notifications: %v", err)
	}
}

func handleSendError(ctx context.Context, b *bot.Bot, err error) {
	if err == nil {
		return
	}
	handleError(ctx, b, errors.Join(ErrorCannotSend, err))
}

func handleSaveError(ctx context.Context, b *bot.Bot, err error) {
	if err == nil {
		return
	}
	handleError(ctx, b, errors.Join(ErrorCannotSave, err))
}

func notifyService(ctx context.Context, b *bot.Bot, msg string) error {
	error_slice := []error{}
	for _, channel := range config.ServiceChannels {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: channel,
			Text:   msg,
		})
		if err != nil {
			error_slice = append(error_slice, err)
		}
	}

	return errors.Join(error_slice...)
}
