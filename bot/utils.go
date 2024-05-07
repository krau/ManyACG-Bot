package bot

import (
	"ManyACG-Bot/service"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/types"
	"context"

	"github.com/mymmrac/telego"
)

func CheckPermissionInGroup(ctx context.Context, message telego.Message, permissions ...types.Permission) bool {
	chatID := message.Chat.ID
	if message.Chat.Type != telego.ChatTypeGroup && message.Chat.Type != telego.ChatTypeSupergroup {
		chatID = message.From.ID
	}
	if !service.CheckAdminPermission(ctx, chatID, permissions...) {
		return service.CheckAdminPermission(ctx, message.From.ID, permissions...)
	}
	return true
}

func MatchSourceURLForMessage(message *telego.Message) string {
	text := message.Text
	text += message.Caption + " "
	for _, entity := range message.Entities {
		if entity.Type == telego.EntityTypeTextLink {
			text += entity.URL + " "
		}
	}
	return sources.FindSourceURL(text)
}
