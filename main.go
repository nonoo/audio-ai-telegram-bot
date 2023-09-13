package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/exp/slices"
)

var telegramBot *bot.Bot
var cmdHandler cmdHandlerType
var reqQueue ReqQueue
var converter Converter
var upload Upload
var tts TTS
var stt STT
var mdx MDX
var rvc RVC
var musicgen Musicgen
var audiogen Audiogen

func sendReplyToMessage(ctx context.Context, replyToMsg *models.Message, s string) (msg *models.Message) {
	var err error
	msg, err = telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ReplyToMessageID: replyToMsg.ID,
		ChatID:           replyToMsg.Chat.ID,
		Text:             s,
	})
	if err != nil {
		fmt.Println("  reply send error:", err)
	}
	return
}

func editReplyToMessage(ctx context.Context, msg *models.Message, s string) error {
	var err error
	_, err = telegramBot.EditMessageText(ctx, &bot.EditMessageTextParams{
		MessageID: msg.ID,
		ChatID:    msg.Chat.ID,
		Text:      s,
	})
	if err != nil {
		fmt.Println("  reply edit error:", err)
	}
	return err
}

func sendTextToAdmins(ctx context.Context, s string) {
	for _, chatID := range params.AdminUserIDs {
		_, _ = telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   s,
		})
	}
}

type AudioFileData struct {
	data     []byte
	filename string
}

func handleAudio(ctx context.Context, update *models.Update, fileID, filename string) {
	// Are we expecting audio data from this user?
	if reqQueue.currentEntry.gotAudioChan == nil || update.Message.From.ID != reqQueue.currentEntry.entry.Message.From.ID {
		return
	}

	var g GetFile
	d, err := g.GetFile(ctx, fileID)
	if err != nil {
		reqQueue.currentEntry.entry.sendReply(ctx, errorStr+": can't get file: "+err.Error())
		return
	}
	reqQueue.currentEntry.entry.sendReply(ctx, doneStr+" downloading\n"+reqQueue.currentEntry.entry.Req.Params.String())
	// Updating the message to reply to this document.
	reqQueue.currentEntry.entry.Message = update.Message
	reqQueue.currentEntry.entry.ReplyMessage = nil
	// Notifying the request queue that we now got the audio data.
	reqQueue.currentEntry.gotAudioChan <- AudioFileData{
		data:     d,
		filename: filename,
	}
}

func handleMessage(ctx context.Context, update *models.Update) {
	fmt.Print("msg from ", update.Message.From.Username, "#", update.Message.From.ID, ": ", update.Message.Text, "\n")

	if update.Message.Chat.ID >= 0 { // From user?
		if !slices.Contains(params.AllowedUserIDs, update.Message.From.ID) {
			fmt.Println("  user not allowed, ignoring")
			return
		}
	} else { // From group ?
		fmt.Print("  msg from group #", update.Message.Chat.ID)
		if !slices.Contains(params.AllowedGroupIDs, update.Message.Chat.ID) {
			fmt.Println(", group not allowed, ignoring")
			return
		}
		fmt.Println()
	}

	// Check if message is a command.
	if update.Message.Text[0] == '/' || update.Message.Text[0] == '!' {
		cmd := strings.Split(update.Message.Text, " ")[0]
		if strings.Contains(cmd, "@") {
			cmd = strings.Split(cmd, "@")[0]
		}
		update.Message.Text = strings.TrimPrefix(update.Message.Text, cmd+" ")
		cmdChar := string(cmd[0])
		cmd = cmd[1:] // Cutting the command character.
		switch cmd {
		case "aaitts":
			fmt.Println("  interpreting as cmd tts")
			cmdHandler.TTS(ctx, strings.Replace(update.Message.Text, cmdChar+"aaitts", "", 1), update.Message)
			return
		case "aaitts-models":
			fmt.Println("  interpreting as cmd tts-models")
			tts.ListModels(ctx, update.Message)
			return
		case "aaistt":
			fmt.Println("  interpreting as cmd stt")
			cmdHandler.STT(ctx, update.Message)
			return
		case "aaimdx":
			fmt.Println("  interpreting as cmd mdx")
			cmdHandler.MDX(ctx, update.Message)
			return
		case "aairvc":
			fmt.Println("  interpreting as cmd rvc")
			cmdHandler.RVC(ctx, update.Message)
			return
		case "aairvc-train":
			fmt.Println("  interpreting as cmd rvc-train")
			cmdHandler.RVCTrain(ctx, update.Message)
			return
		case "aairvc-models":
			fmt.Println("  interpreting as cmd rvc")
			rvc.ListModels(ctx, update.Message)
			return
		case "aaimusicgen":
			fmt.Println("  interpreting as cmd musicgen")
			cmdHandler.Musicgen(ctx, strings.Replace(update.Message.Text, cmdChar+"aaimusicgen", "", 1), update.Message)
			return
		case "aaiaudiogen":
			fmt.Println("  interpreting as cmd audiogen")
			cmdHandler.Audiogen(ctx, strings.Replace(update.Message.Text, cmdChar+"aaiaudiogen", "", 1), update.Message)
			return
		case "aaicancel":
			fmt.Println("  interpreting as cmd aaicancel")
			cmdHandler.Cancel(ctx, update.Message)
			return
		case "aaihelp":
			fmt.Println("  interpreting as cmd aaihelp")
			cmdHandler.Help(ctx, update.Message, cmdChar)
			return
		case "start":
			fmt.Println("  interpreting as cmd start")
			if update.Message.Chat.ID >= 0 { // From user?
				sendReplyToMessage(ctx, update.Message, "🤖 Welcome! This is the Audio AI Telegram Bot for "+
					"processing audio content with AI.\n\nMore info: https://github.com/nonoo/audio-ai-telegram-bot")
			}
			return
		default:
			fmt.Println("  invalid cmd")
			if update.Message.Chat.ID >= 0 {
				sendReplyToMessage(ctx, update.Message, errorStr+": invalid command")
			}
			return
		}
	}

	if update.Message.Chat.ID >= 0 { // From user?
		cmdHandler.TTS(ctx, update.Message.Text, update.Message)
	}
}

func telegramBotUpdateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if update.Message.Document != nil {
		handleAudio(ctx, update, update.Message.Document.FileID, update.Message.Document.FileName)
	} else if update.Message.Voice != nil {
		handleAudio(ctx, update, update.Message.Voice.FileID, "voice.ogg")
	} else if update.Message.Audio != nil {
		handleAudio(ctx, update, update.Message.Audio.FileID, update.Message.Audio.FileName)
	} else if update.Message != nil && update.Message.Text != "" {
		handleMessage(ctx, update)
	}
}

func main() {
	fmt.Println("audio-ai-telegram-bot starting...")

	if err := params.Init(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	var cancel context.CancelFunc
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	reqQueue.Init(ctx)

	opts := []bot.Option{
		bot.WithDefaultHandler(telegramBotUpdateHandler),
	}

	var err error
	telegramBot, err = bot.New(params.BotToken, opts...)
	if nil != err {
		panic(fmt.Sprint("can't init telegram bot: ", err))
	}

	sendTextToAdmins(ctx, "🤖 Bot started")

	telegramBot.Start(ctx)
}
