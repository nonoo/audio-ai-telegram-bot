package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-telegram/bot/models"
)

type TTS struct {
}

var TTSOutFilePath = os.TempDir() + "/tts.wav"

func (t *TTS) ListModels(ctx context.Context, msg *models.Message) {
	msg = sendReplyToMessage(ctx, msg, "👅 Querying...")
	cmd := exec.Command(params.TTSBin, "--list_models")
	cmd.Dir = path.Dir(params.TTSBin)
	output, err := cmd.Output()
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't list models: "+err.Error())
		return
	}
	_ = editReplyToMessage(ctx, msg, "👅 Available models:\n\n"+string(output))
}

func (t *TTS) CleanupOutputFiles() {
	os.Remove(TTSOutFilePath)
}

func (t *TTS) TTS(ctx context.Context, reqParams ReqParamsTTS, prompt string) (io.ReadCloser, error) {
	t.CleanupOutputFiles()

	cmd := NewCommand(ctx, params.TTSBin, "--model_name", reqParams.Model, "--out_path", TTSOutFilePath)
	cmd.Dir = path.Dir(params.TTSBin)
	cmd.Stdin = strings.NewReader(prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.CleanupOutputFiles()
		return nil, fmt.Errorf("TTS error: %w: %s", err, string(output))
	}

	// Check output .wav file
	if stat, err := os.Stat(TTSOutFilePath); os.IsNotExist(err) || stat.Size() == 0 {
		t.CleanupOutputFiles()
		return nil, fmt.Errorf("output file not found: %s", TTSOutFilePath)
	}

	r, err := converter.ConvertToOpus(ctx, TTSOutFilePath)
	if err != nil {
		t.CleanupOutputFiles()
		return nil, err
	}

	return r, nil
}
