package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Upload struct {
}

func (u *Upload) Voice(ctx context.Context, qEntry *ReqQueueEntry, r io.ReadCloser, retryAllowed bool) error {
	defer r.Close()

	fmt.Println("  uploading...")
	qEntry.sendUpdate(ctx, uploadingStr)

	params := &bot.SendVoiceParams{
		ChatID:           qEntry.Message.Chat.ID,
		ReplyToMessageID: qEntry.Message.ID,
		Voice: &models.InputFileUpload{
			Filename: "tts-" + fmt.Sprint(qEntry.TaskID) + ".ogg",
			Data:     r,
		},
		// Caption: qEntry.Req.Message.Text,
	}
	_, err := telegramBot.SendVoice(ctx, params)
	if err != nil {
		fmt.Println("  send error:", err)

		if !retryAllowed {
			return fmt.Errorf("send error: %w", err)
		}

		retryAfter := qEntry.checkWaitError(err)
		if retryAfter > 0 {
			fmt.Println("  retrying after", retryAfter, "...")
			time.Sleep(retryAfter)
			return u.Voice(ctx, qEntry, r, false)
		}
		return err
	}

	return nil
}

type UploadFileData struct {
	r        io.ReadCloser
	filename string
}

func (u *Upload) Audio(ctx context.Context, qEntry *ReqQueueEntry, f []UploadFileData, retryAllowed bool) error {
	defer func() {
		for i := range f {
			f[i].r.Close()
		}
	}()

	fmt.Println("  uploading...")
	qEntry.sendUpdate(ctx, uploadingStr)

	var media []models.InputMedia
	for i := range f {
		media = append(media, &models.InputMediaAudio{
			Media:           "attach://" + f[i].filename,
			MediaAttachment: f[i].r,
		})
	}
	params := &bot.SendMediaGroupParams{
		ChatID:           qEntry.Message.Chat.ID,
		ReplyToMessageID: qEntry.Message.ID,
		Media:            media,
	}
	_, err := telegramBot.SendMediaGroup(ctx, params)
	if err != nil {
		fmt.Println("  send error:", err)

		if !retryAllowed {
			return fmt.Errorf("send error: %w", err)
		}

		retryAfter := qEntry.checkWaitError(err)
		if retryAfter > 0 {
			fmt.Println("  retrying after", retryAfter, "...")
			time.Sleep(retryAfter)
			return u.Audio(ctx, qEntry, f, false)
		}
		return err
	}

	return nil
}
