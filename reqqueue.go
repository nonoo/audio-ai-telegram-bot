package main

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot/models"
)

const audioReqStr = "🎙️ Please post the audio file to process."
const processStr = "🔨 Processing"
const progressBarLength = 20
const downloadingStr = "⬇ Downloading..."
const uploadingStr = "☁ ️ Uploading..."
const doneStr = "✅ Done"
const errorStr = "❌ Error"
const canceledStr = "❌ Canceled"

const processTimeout = 5 * time.Minute
const groupChatProgressUpdateInterval = 3 * time.Second
const privateChatProgressUpdateInterval = 500 * time.Millisecond

type ReqType int

const (
	ReqTypeTTS ReqType = iota
	ReqTypeSTT
	ReqTypeMDX
	ReqTypeRVC
	ReqTypeMusicgen
	ReqTypeAudiogen
)

type ReqQueueEntry struct {
	TaskID uint64

	ReplyMessage *models.Message
	Message      *models.Message
	Req          ReqQueueReq

	LastProcessUpdateAt time.Time
	ProcessUpdateTimer  *time.Timer
}

func (e *ReqQueueEntry) checkWaitError(err error) time.Duration {
	var retryRegex = regexp.MustCompile(`{"retry_after":([0-9]+)}`)
	match := retryRegex.FindStringSubmatch(err.Error())
	if len(match) < 2 {
		return 0
	}

	retryAfter, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}
	return time.Duration(retryAfter) * time.Second
}

func (e *ReqQueueEntry) sendReply(ctx context.Context, s string) {
	if e.ReplyMessage == nil {
		e.ReplyMessage = sendReplyToMessage(ctx, e.Message, s)
	} else if e.ReplyMessage.Text != s {
		e.ReplyMessage.Text = s
		err := editReplyToMessage(ctx, e.ReplyMessage, s)
		if err != nil {
			waitNeeded := e.checkWaitError(err)
			fmt.Println("  waiting", waitNeeded, "...")
			time.Sleep(waitNeeded)
		}
	}
}

func (e *ReqQueueEntry) cancelProcessUpdate() {
	if e.ProcessUpdateTimer != nil {
		e.ProcessUpdateTimer.Stop()
		select {
		case <-e.ProcessUpdateTimer.C:
		default:
		}
		e.ProcessUpdateTimer = nil
	}
}

func (e *ReqQueueEntry) sendProcessUpdate(ctx context.Context, processDesc string, percent int) {
	e.cancelProcessUpdate()
	updateInterval := groupChatProgressUpdateInterval
	if e.Message.Chat.ID > 0 {
		updateInterval = privateChatProgressUpdateInterval
	}
	elapsedSinceLastUpdate := time.Since(e.LastProcessUpdateAt)
	if elapsedSinceLastUpdate < updateInterval {
		e.ProcessUpdateTimer = time.AfterFunc(updateInterval-elapsedSinceLastUpdate, func() {
			e.sendProcessUpdate(ctx, processDesc, percent)
		})
		return
	}

	e.LastProcessUpdateAt = time.Now()

	var str string
	if processDesc == "" {
		str = processStr
	} else {
		str = processDesc
	}
	if percent >= 0 {
		str += " " + getProgressbar(percent, progressBarLength)
	} else {
		if !strings.HasSuffix(str, "...") {
			str += "..."
		}
	}
	reqParamsStr := e.Req.Params.String()
	if len(reqParamsStr) > 0 {
		str += "\n" + reqParamsStr
	}
	e.sendReply(ctx, str)
}

func (e *ReqQueueEntry) sendUpdate(ctx context.Context, s string) {
	reqParamsStr := e.Req.Params.String()
	if len(reqParamsStr) > 0 {
		s += "\n" + reqParamsStr
	}
	e.sendReply(ctx, s)
}

// func (e *ReqQueueEntry) deleteReply(ctx context.Context) {
// 	if e.ReplyMessage == nil {
// 		return
// 	}

// 	_, _ = telegramBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
// 		MessageID: e.ReplyMessage.ID,
// 		ChatID:    e.ReplyMessage.Chat.ID,
// 	})
// }

type ReqQueueReq struct {
	Type    ReqType
	Message *models.Message
	Prompt  string
	Params  ReqParams
}

type ReqQueueCurrentEntry struct {
	entry     *ReqQueueEntry
	canceled  bool
	ctxCancel context.CancelFunc

	gotAudioChan chan AudioFileData
}

type ReqQueue struct {
	mutex          sync.Mutex
	ctx            context.Context
	entries        []ReqQueueEntry
	processReqChan chan bool

	currentEntry ReqQueueCurrentEntry
}

func (q *ReqQueue) Add(req ReqQueueReq) {
	q.mutex.Lock()

	newEntry := ReqQueueEntry{
		TaskID:  rand.Uint64(),
		Message: req.Message,
		Req:     req,
	}

	if len(q.entries) > 0 {
		fmt.Println("  queueing request at position #", len(q.entries))
		newEntry.sendReply(q.ctx, q.getQueuePositionString(len(q.entries)))
	}

	q.entries = append(q.entries, newEntry)
	q.mutex.Unlock()

	select {
	case q.processReqChan <- true:
	default:
	}
}

func (q *ReqQueue) CancelCurrentEntry(ctx context.Context) (err error) {
	q.mutex.Lock()
	if len(q.entries) > 0 {
		fmt.Println("  cancelling active request")
		q.currentEntry.canceled = true
		q.currentEntry.ctxCancel()
	} else {
		fmt.Println("  no active request to cancel")
		err = fmt.Errorf("no active request to cancel")
	}
	q.mutex.Unlock()
	return
}

func (q *ReqQueue) getQueuePositionString(pos int) string {
	return "👨‍👦‍👦 Request queued at position #" + fmt.Sprint(pos)
}

func (q *ReqQueue) processQueueEntry(processCtx context.Context, qEntry *ReqQueueEntry, audioData AudioFileData) error {
	fmt.Print("processing request from ", q.currentEntry.entry.Message.From.Username, "#", q.currentEntry.entry.Message.From.ID,
		": ", q.currentEntry.entry.Req.Message.Text, "\n")

	qEntry.sendProcessUpdate(q.ctx, "", -1)

	switch qEntry.Req.Type {
	case ReqTypeTTS:
		reader, err := tts.TTS(processCtx, qEntry.Req.Params.(ReqParamsTTS), qEntry.Req.Prompt)
		if err != nil {
			return err
		}

		err = upload.Voice(q.ctx, q.currentEntry.entry, reader, true)
		if err != nil {
			return err
		}

		q.currentEntry.entry.sendUpdate(q.ctx, doneStr)
	case ReqTypeSTT:
		text, err := stt.STT(processCtx, qEntry.Req.Params.(ReqParamsSTT), audioData)
		if err != nil {
			return err
		}

		fmt.Println("  result:", text)
		q.currentEntry.entry.sendReply(q.ctx, text)
	case ReqTypeMDX:
		files, err := mdx.MDX(processCtx, qEntry.Req.Params.(ReqParamsMDX), audioData)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("got no output files from MDX")
		}

		defer mdx.CleanupOutputFiles()

		err = upload.Audio(q.ctx, q.currentEntry.entry, files, true)
		if err != nil {
			return err
		}

		q.currentEntry.entry.sendUpdate(q.ctx, doneStr)
	case ReqTypeRVC:
		file, err := rvc.RVC(processCtx, qEntry.Req.Params.(ReqParamsRVC), audioData)
		if err != nil {
			return err
		}

		defer rvc.CleanupOutputFiles()

		err = upload.Voice(q.ctx, q.currentEntry.entry, file, true)
		if err != nil {
			return err
		}

		q.currentEntry.entry.sendUpdate(q.ctx, doneStr)
	case ReqTypeMusicgen:
		file, err := musicgen.Musicgen(processCtx, qEntry.Req.Params.(ReqParamsMusicgen), qEntry.Req.Prompt, audioData)
		if err != nil {
			return err
		}

		defer musicgen.CleanupOutputFiles()

		err = upload.Voice(q.ctx, q.currentEntry.entry, file, true)
		if err != nil {
			return err
		}

		q.currentEntry.entry.sendUpdate(q.ctx, doneStr)
	case ReqTypeAudiogen:
		file, err := audiogen.Audiogen(processCtx, qEntry.Req.Params.(ReqParamsAudiogen), qEntry.Req.Prompt)
		if err != nil {
			return err
		}

		defer audiogen.CleanupOutputFiles()

		err = upload.Voice(q.ctx, q.currentEntry.entry, file, true)
		if err != nil {
			return err
		}

		q.currentEntry.entry.sendUpdate(q.ctx, doneStr)
	}

	return nil
}

func (q *ReqQueue) processor() {
	for {
		q.mutex.Lock()
		if (len(q.entries)) == 0 {
			q.mutex.Unlock()
			<-q.processReqChan
			continue
		}

		// Updating queue positions for all waiting entries.
		for i := 1; i < len(q.entries); i++ {
			sendReplyToMessage(q.ctx, q.entries[i].Message, q.getQueuePositionString(i))
		}

		q.currentEntry = ReqQueueCurrentEntry{}
		var processCtx context.Context
		processCtx, q.currentEntry.ctxCancel = context.WithTimeout(q.ctx, processTimeout)
		q.currentEntry.entry = &q.entries[0]
		q.mutex.Unlock()

		reqParamsStr := q.currentEntry.entry.Req.Params.String()
		if len(reqParamsStr) > 0 {
			fmt.Println("  request params:", reqParamsStr)
		}

		var err error
		var audioData AudioFileData
		audioNeededFirst := false
		switch q.currentEntry.entry.Req.Type {
		case ReqTypeSTT, ReqTypeMDX, ReqTypeRVC, ReqTypeMusicgen:
			audioNeededFirst = true
		}
		if audioNeededFirst {
			fmt.Println("  waiting for audio file...")
			q.currentEntry.entry.sendUpdate(q.ctx, audioReqStr)
			q.currentEntry.gotAudioChan = make(chan AudioFileData)
			select {
			case audioData = <-q.currentEntry.gotAudioChan:
			case <-processCtx.Done():
				q.currentEntry.canceled = true
			case <-time.NewTimer(3 * time.Minute).C:
				fmt.Println("  waiting for audio file timeout")
				err = fmt.Errorf("waiting for audio data timeout")
			}
			close(q.currentEntry.gotAudioChan)
			q.currentEntry.gotAudioChan = nil

			if err == nil && len(audioData.data) == 0 {
				err = fmt.Errorf("got no audio data")
			}
		}

		if err == nil {
			err = q.processQueueEntry(processCtx, q.currentEntry.entry, audioData)
		}

		q.mutex.Lock()
		if q.currentEntry.canceled {
			fmt.Print("  canceled\n")
			q.currentEntry.entry.sendReply(q.ctx, canceledStr)
		} else if err != nil {
			fmt.Println("  error:", err)
			q.currentEntry.entry.sendReply(q.ctx, errorStr+": "+err.Error())
		}

		q.currentEntry.ctxCancel()

		q.entries = q.entries[1:]
		if len(q.entries) == 0 {
			fmt.Print("finished queue processing\n")
		}
		q.mutex.Unlock()
	}
}

func (q *ReqQueue) Init(ctx context.Context) {
	q.ctx = ctx
	q.processReqChan = make(chan bool)
	go q.processor()
}
