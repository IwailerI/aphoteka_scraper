package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

//go:embed token.secret
var token string

//go:embed chats.secret
var channels_raw string

var channel_ids []string

func init() {
	parts := strings.Split(channels_raw, "\n")
	for _, part := range parts {
		channel_ids = append(channel_ids, strings.TrimSpace(part))
	}
}

func generate_icon(tag string) string {
	switch tag {
	case "https://schema.org/OutOfStock":
		return "❌"
	case "https://schema.org/InStock":
		return "✅"
	default:
		return "⚠️"
	}
}

func generate_message(input map[string]Availability) string {
	var builder strings.Builder

	keys := []string{}
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, name := range keys {
		availability := input[name]
		if availability.tag == "" {
			fmt.Fprintf(&builder, "- ❌ %v: not found\n%v\n\n", name, availability.url)
		} else {
			parts := strings.Split(availability.tag, "/")
			fmt.Fprintf(
				&builder, "- %s %v: %v @ %.2f %s\n%v\n\n",
				generate_icon(availability.tag), name, parts[len(parts)-1],
				float64(availability.price)*0.01, availability.currency, availability.url,
			)
		}
	}

	return strings.TrimSpace(builder.String())
}

func send_update(msg string) error {
	log.Printf("Updating %d channels", len(channel_ids))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := bot.New(token)
	if err != nil {
		return err
	}

	good := 0

	for _, channel := range channel_ids {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: channel,
			Text:   msg,
			LinkPreviewOptions: &models.LinkPreviewOptions{
				IsDisabled: bot.True(),
			},
		})
		if err != nil {
			log.Printf("Cannot send message in channel %v: %v", channel, err)
		} else {
			good++
		}
	}

	log.Printf("Updated %d channels", good)

	return nil
}
