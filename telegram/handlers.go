package telegram

import (
	"ManyACG-Bot/service"
	"ManyACG-Bot/types"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"

	. "ManyACG-Bot/logger"
)

func start(ctx context.Context, bot *telego.Bot, message telego.Message) {
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > 0 {
		Logger.Debugf("start: args=%v", args)
		if strings.HasPrefix(args[0], "file_") {
			messageIDStr := args[0][5:]
			messageID, err := strconv.Atoi(messageIDStr)
			if err != nil {
				bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取失败: %s", err).WithReplyParameters(
					&telego.ReplyParameters{
						MessageID: message.MessageID,
					},
				))
				return
			}
			_, err = sendPictureFileByMessageID(ctx, bot, message, messageID)
			if err != nil {
				bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取失败: %s", err).WithReplyParameters(
					&telego.ReplyParameters{
						MessageID: message.MessageID,
					},
				))
				return
			}
		}
		return
	}

	bot.SendMessage(
		telegoutil.Message(message.Chat.ChatID(),
			"Hi~").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
}

func getPictureFile(ctx context.Context, bot *telego.Bot, message telego.Message) {
	replyToMessage := message.ReplyToMessage
	if replyToMessage == nil {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请使用该命令回复一条频道的消息").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if replyToMessage.Photo == nil {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "目标消息不包含图片").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if replyToMessage.ForwardOrigin == nil {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请使用该命令回复一条频道的消息").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}

	var messageOriginChannel *telego.MessageOriginChannel
	if replyToMessage.ForwardOrigin.OriginType() == telego.OriginTypeChannel {
		messageOriginChannel = replyToMessage.ForwardOrigin.(*telego.MessageOriginChannel)
	} else {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请使用该命令回复一条频道的消息").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}

	_, err := sendPictureFileByMessageID(ctx, bot, message, messageOriginChannel.MessageID)
	if err != nil {
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取失败: %s", err).WithReplyParameters(
			&telego.ReplyParameters{
				MessageID: message.MessageID,
			},
		))
		return
	}
}

func randomPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	r18 := cmd == "setu"
	limit := 1
	Logger.Debugf("randomPicture: r18=%v, args=%v", r18, args)
	artwork, err := service.GetRandomArtworksByTagsR18(ctx, args, r18, limit)
	if err != nil {
		Logger.Warnf("获取图片失败: %s", err)
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取图片失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if len(artwork) == 0 {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "未找到图片").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	picture := artwork[0].Pictures[0]
	var file telego.InputFile
	if picture.TelegramInfo.PhotoFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
	} else {
		photoURL := picture.Original
		if artwork[0].SourceType == types.SourceTypePixiv {
			photoURL = strings.Replace(photoURL, "img-original", "img-master", 1)
			photoURL = strings.Replace(photoURL, ".jpg", "_master1200.jpg", 1)
			photoURL = strings.Replace(photoURL, ".png", "_master1200.jpg", 1)
		}
		file = telegoutil.FileFromURL(photoURL)
	}
	_, err = bot.SendPhoto(telegoutil.Photo(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption(artwork[0].Title).WithReplyMarkup(
		telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("来源").WithURL(fmt.Sprintf("https://t.me/%s/%d", strings.ReplaceAll(ChannelChatID.String(), "@", ""), picture.TelegramInfo.MessageID)),
			telegoutil.InlineKeyboardButton("原图").WithURL(fmt.Sprintf("https://t.me/%s/?start=file_%d", BotUsername, picture.TelegramInfo.MessageID)),
		}),
	))
	if err != nil {
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "发送图片失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
	}

}
