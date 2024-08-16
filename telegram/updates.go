package telegram

import (
	_ "embed"
)

// func SendUpdate(msg string) error {
// 	msg = cleanMsg(msg)

// 	log.Printf("Updating %d channels", len(secrets.ChannelIds))

// 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
// 	defer cancel()

// 	b, err := bot.New(secrets.Token)
// 	if err != nil {
// 		return err
// 	}

// 	good := 0

// 	for _, channel := range secrets.ChannelIds {
// 		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
// 			ChatID: channel,
// 			Text:   msg,
// 			LinkPreviewOptions: &models.LinkPreviewOptions{
// 				IsDisabled: bot.True(),
// 			},
// 		})
// 		if err != nil {
// 			log.Printf("Cannot send message in channel %v: %v", channel, err)
// 		} else {
// 			good++
// 		}
// 	}

// 	log.Printf("Updated %d channels", good)
// 	permanence.Logger.TelegramNotified()

// 	return nil
// }

// func SendErrors() error {
// 	e := errors.Join(permanence.Logger.GetErrors()...)
// 	msg := cleanMsg("Errors when executing!\n" + e.Error())

// 	err := SendUpdate(msg)
// 	if err != nil {
// 		permanence.Logger.AddError(err)
// 		return err
// 	}

// 	return nil
// }

// func cleanMsg(msg string) string {
// 	msg = strings.ReplaceAll(msg, secrets.Token, "***")

// 	for _, chat := range secrets.ChannelIds {
// 		msg = strings.ReplaceAll(msg, chat, "***")
// 	}

// 	return msg

// }
