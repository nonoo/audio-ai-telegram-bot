package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type paramsType struct {
	BotToken string

	AllowedUserIDs  []int64
	AdminUserIDs    []int64
	AllowedGroupIDs []int64

	TTSBin          string
	TTSDefaultModel string

	STTBin string

	MDXBin string

	RVCBin          string
	RVCModelPath    string
	RVCDefaultModel string

	RVCTrainBin              string
	RVCTrainDefaultBatchSize int
	RVCTrainDefaultEpochs    int

	MusicgenBin string

	AudiogenBin string
}

var params paramsType

func (p *paramsType) Init() error {
	flag.StringVar(&p.BotToken, "bot-token", "", "telegram bot token")
	var allowedUserIDs string
	flag.StringVar(&allowedUserIDs, "allowed-user-ids", "", "allowed telegram user ids")
	var adminUserIDs string
	flag.StringVar(&adminUserIDs, "admin-user-ids", "", "admin telegram user ids")
	var allowedGroupIDs string
	flag.StringVar(&allowedGroupIDs, "allowed-group-ids", "", "allowed telegram group ids")
	flag.StringVar(&p.TTSBin, "tts-bin", "", "path to the tts binary")
	flag.StringVar(&p.TTSDefaultModel, "tts-default-model", "", "default tts model")
	flag.StringVar(&p.STTBin, "stt-bin", "", "path to the stt binary")
	flag.StringVar(&p.MDXBin, "mdx-bin", "", "path to the mdx binary")
	flag.StringVar(&p.RVCBin, "rvc-bin", "", "path to the rvc binary")
	flag.StringVar(&p.RVCModelPath, "rvc-model-path", "", "path to the rvc weights directory")
	flag.StringVar(&p.RVCDefaultModel, "rvc-default-model", "", "default rvc model")
	flag.StringVar(&p.RVCTrainBin, "rvc-train-bin", "", "path to the rvc train binary")
	flag.IntVar(&p.RVCTrainDefaultBatchSize, "rvc-train-default-batch-size", 0, "default rvc train batch size")
	flag.IntVar(&p.RVCTrainDefaultEpochs, "rvc-train-default-epochs", 0, "default rvc train epochs")
	flag.StringVar(&p.MusicgenBin, "musicgen-bin", "", "path to the musicgen binary")
	flag.StringVar(&p.AudiogenBin, "audiogen-bin", "", "path to the audiogen binary")
	flag.Parse()

	if p.BotToken == "" {
		p.BotToken = os.Getenv("BOT_TOKEN")
	}
	if p.BotToken == "" {
		return fmt.Errorf("bot token not set")
	}

	if allowedUserIDs == "" {
		allowedUserIDs = os.Getenv("ALLOWED_USERIDS")
	}
	sa := strings.Split(allowedUserIDs, ",")
	for _, idStr := range sa {
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("allowed user ids contains invalid user ID: " + idStr)
		}
		p.AllowedUserIDs = append(p.AllowedUserIDs, id)
	}

	if adminUserIDs == "" {
		adminUserIDs = os.Getenv("ADMIN_USERIDS")
	}
	sa = strings.Split(adminUserIDs, ",")
	for _, idStr := range sa {
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("admin ids contains invalid user ID: " + idStr)
		}
		p.AdminUserIDs = append(p.AdminUserIDs, id)
		if !slices.Contains(p.AllowedUserIDs, id) {
			p.AllowedUserIDs = append(p.AllowedUserIDs, id)
		}
	}

	if allowedGroupIDs == "" {
		allowedGroupIDs = os.Getenv("ALLOWED_GROUPIDS")
	}
	sa = strings.Split(allowedGroupIDs, ",")
	for _, idStr := range sa {
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("allowed group ids contains invalid group ID: " + idStr)
		}
		p.AllowedGroupIDs = append(p.AllowedGroupIDs, id)
	}

	if p.TTSBin == "" {
		p.TTSBin = os.Getenv("TTS_BIN")
	}

	if p.TTSDefaultModel == "" {
		p.TTSDefaultModel = os.Getenv("TTS_DEFAULT_MODEL")
	}

	if p.STTBin == "" {
		p.STTBin = os.Getenv("STT_BIN")
	}

	if p.MDXBin == "" {
		p.MDXBin = os.Getenv("MDX_BIN")
	}

	if p.RVCBin == "" {
		p.RVCBin = os.Getenv("RVC_BIN")
	}
	if p.RVCModelPath == "" {
		p.RVCModelPath = os.Getenv("RVC_MODEL_PATH")
	}
	if p.RVCDefaultModel == "" {
		p.RVCDefaultModel = os.Getenv("RVC_DEFAULT_MODEL")
	}

	if p.RVCTrainBin == "" {
		p.RVCTrainBin = os.Getenv("RVC_TRAIN_BIN")
	}
	if p.RVCTrainDefaultBatchSize == 0 {
		p.RVCTrainDefaultBatchSize, _ = strconv.Atoi(os.Getenv("RVC_TRAIN_DEFAULT_BATCH_SIZE"))
	}
	if p.RVCTrainDefaultEpochs == 0 {
		p.RVCTrainDefaultEpochs, _ = strconv.Atoi(os.Getenv("RVC_TRAIN_DEFAULT_EPOCHS"))
	}

	if p.MusicgenBin == "" {
		p.MusicgenBin = os.Getenv("MUSICGEN_BIN")
	}

	if p.AudiogenBin == "" {
		p.AudiogenBin = os.Getenv("AUDIOGEN_BIN")
	}

	return nil
}
