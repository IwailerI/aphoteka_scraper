package secrets

import (
	_ "embed"
	"strings"
)

//go:embed token.secret
var Token string

//go:embed chats.secret
var channelsRaw string

var ChannelIds []string

func init() {
	parts := strings.Split(channelsRaw, "\n")
	for _, part := range parts {
		ChannelIds = append(ChannelIds, strings.TrimSpace(part))
	}
}
