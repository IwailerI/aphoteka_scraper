package telegram

import (
	"aphoteka_scraper/permanence"
	"aphoteka_scraper/secrets"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func handlerAddUser(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	username, ok := strings.CutPrefix(update.Message.Text, "/add_user ")
	username = "@" + strings.TrimPrefix(strings.TrimSpace(username), "@")

	if !ok || len(username) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_user <username>",
		})
		handleSendError(ctx, b, err)
		return
	}

	config.Whitelist[username] = unit
	err := saveServerConfig()
	handleSaveError(ctx, b, err)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("User %q added.", username),
	})
	handleSendError(ctx, b, err)
}

func handleRemoveUser(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	username, ok := strings.CutPrefix(update.Message.Text, "/remove_user ")
	username = "@" + strings.TrimPrefix(strings.TrimSpace(username), "@")

	if !ok || len(username) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_user <username>",
		})
		handleSendError(ctx, b, err)
		return
	}

	if username == secrets.RootUser {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Cannot remove root user. (yay, you found who is root)",
		})
		handleSendError(ctx, b, err)
		return
	}

	_, found := config.Whitelist[username]
	if !found {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("User %q is not in whitelist.", username),
		})
		handleSendError(ctx, b, err)
		return
	}

	delete(config.Whitelist, username)
	err := saveServerConfig()
	handleSaveError(ctx, b, err)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("User %q removed.", username),
	})
	handleSaveError(ctx, b, err)
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

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Whitelist: %q", keys),
	})
	handleSendError(ctx, b, err)
}

func handleAddChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/add_channel ")
	channel = strings.TrimSpace(channel)

	if !ok || len(channel) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_channel <channel>",
		})
		handleSendError(ctx, b, err)
		return
	}

	if !slices.Contains(config.NotifyChannels, channel) {
		config.NotifyChannels = append(config.NotifyChannels, channel)
	}
	err := saveServerConfig()
	handleSaveError(ctx, b, err)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Channel %q added.", channel),
	})
	handleSendError(ctx, b, err)
}

func handleRemoveChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/remove_channel")
	channel = strings.TrimSpace(channel)

	if !ok || len(channel) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_channel <channel>",
		})
		handleSendError(ctx, b, err)
		return
	}

	i := slices.Index(config.NotifyChannels, channel)
	if i == -1 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Channel %q is not in found.", channel),
		})
		handleSendError(ctx, b, err)
		return
	}

	config.NotifyChannels = swapRemove(config.NotifyChannels, i)
	err := saveServerConfig()
	handleSaveError(ctx, b, err)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Channel %q will not be notified anymore.", channel),
	})
	handleSendError(ctx, b, err)

}

func handleAddServiceChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/add_service_channel ")

	if !ok || len(channel) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_service_channel <channel>",
		})
		handleSendError(ctx, b, err)
		return
	}

	if !slices.Contains(config.ServiceChannels, channel) {
		config.ServiceChannels = append(config.ServiceChannels, channel)
	}
	err := saveServerConfig()
	handleSaveError(ctx, b, err)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Service channel %q added.", channel),
	})
	handleSendError(ctx, b, err)
}

func handleRemoveServiceChannel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	channel, ok := strings.CutPrefix(update.Message.Text, "/remove_service_channel ")

	if !ok || len(channel) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_service_channel <channel>",
		})
		handleSendError(ctx, b, err)
		return
	}

	i := slices.Index(config.ServiceChannels, channel)
	if i == -1 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Channel %q is not in found.", channel),
		})
		handleSendError(ctx, b, err)
		return
	}

	config.ServiceChannels = swapRemove(config.ServiceChannels, i)
	err := saveServerConfig()
	handleSaveError(ctx, b, err)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Service channel %q will not be notified anymore.", channel),
	})
	handleSendError(ctx, b, err)
}

func handleListChannels(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	sort.Strings(config.NotifyChannels)
	sort.Strings(config.ServiceChannels)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Notify: %q\nService: %q", config.NotifyChannels, config.ServiceChannels),
	})
	handleSendError(ctx, b, err)
}

func handleForceUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Updating",
	})
	handleSendError(ctx, b, err)

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
		handleSaveError(ctx, b, err)
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Update cycle started.",
	})
	handleSendError(ctx, b, err)
}

func handleStopUpdates(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	if config.Active {
		loopStopHandle <- unit
		config.Active = false
		err := saveServerConfig()
		handleSaveError(ctx, b, err)
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Update cycle stopped.",
	})
	handleSendError(ctx, b, err)
}

func handleSetUpdateInterval(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	s, ok := strings.CutPrefix(update.Message.Text, "/set_update_interval ")
	if !ok {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /set_update_interval <number of minutes>",
		})
		handleSendError(ctx, b, err)
		return
	}
	s = strings.TrimSpace(s)

	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Expected positive integer as first argument",
		})
		handleSendError(ctx, b, err)
		return
	}

	config.Interval = time.Duration(n) * time.Minute
	err = saveServerConfig()
	handleSaveError(ctx, b, err)
	setupLoop(ctx, b)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Interval updated to %d hours %d minutes.", config.Interval/time.Hour, config.Interval%time.Hour/time.Minute),
	})
	handleSendError(ctx, b, err)
}

func handleAddProduct(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	s, ok := strings.CutPrefix(update.Message.Text, "/add_product ")
	slice := strings.Fields(s)
	if !ok || len(slice) != 2 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /add_product <name_of_product> <url>",
		})
		handleSendError(ctx, b, err)
		return
	}

	prev_url, overridden := config.Products[slice[0]]

	if !(overridden && prev_url == slice[1]) {
		config.Products[slice[0]] = slice[1]

		err := saveServerConfig()
		handleSaveError(ctx, b, err)
	}

	if overridden {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: fmt.Sprintf("Product %q now has url %q instead of %q.",
				slice[0],
				slice[1],
				prev_url,
			),
		})
		handleSendError(ctx, b, err)
	} else {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: fmt.Sprintf("Product %q added with url %q.",
				slice[0],
				slice[1],
			),
		})
		handleSendError(ctx, b, err)
	}
}

func handleRemoveProduct(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	s, ok := strings.CutPrefix(update.Message.Text, "/remove_product ")
	s = strings.TrimSpace(s)
	if !ok || s == "" {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Syntax: /remove_product <name_of_product>",
		})
		handleSendError(ctx, b, err)
		return
	}

	_, found := config.Products[s]
	if !found {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Product %q is not found.", s),
		})
		handleSendError(ctx, b, err)
		return
	}
	delete(config.Products, s)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Product %q is deleted.", s),
	})
	handleSendError(ctx, b, err)
}

func handleListProducts(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	if len(config.Products) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "There are no products tracked.",
		})
		handleSendError(ctx, b, err)
		return
	}

	var s strings.Builder

	for product, url := range config.Products {
		fmt.Fprintf(&s, "%s - %s\n", product, url)
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   s.String(),
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
	})
	handleSendError(ctx, b, err)
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
		handleError(ctx, b, errors.Join(ErrorCannotDumpManifest, err))
	}

	fmt.Fprintf(&s, "Human readable interval: %v\n", config.Interval)

	lastManifest, err := permanence.LoadManifest()
	if err == nil {
		fmt.Fprintf(&s, "Last manifest:\n%s\n\n", bot.EscapeMarkdown(lastManifest.GenerateMessage()))
	} else {
		handleError(ctx, b, errors.Join(ErrorCannotLoadManifest, err))
	}

	if !lastCheck.IsZero() {
		fmt.Fprintf(&s, "Last check:\n`%v`\n",
			bot.EscapeMarkdown(fmt.Sprint(lastCheck)))
	}

	if !nextCheck.IsZero() && config.Active {
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

	log.Printf("Status command output:\n%s", s.String())

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      s.String(),
		ParseMode: models.ParseModeMarkdown,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
	})
	handleSendError(ctx, b, err)
}

func handleCheckNow(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !checkPermission(ctx, b, update) {
		return
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Updating",
	})
	handleSendError(ctx, b, err)

	checkAndNotify(ctx, b, false)
}
