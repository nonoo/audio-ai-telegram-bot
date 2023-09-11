package main

import (
	"context"
	"fmt"
	"os"
	"path"
)

type STT struct {
}

var STTInFilePath = os.TempDir() + "/tts.wav"
var STTOutFilePath = os.TempDir() + "/tts.txt"

func (t *STT) STT(ctx context.Context, reqParams ReqParamsSTT, audioData AudioFileData) (string, error) {
	os.Remove(STTInFilePath)
	os.Remove(STTOutFilePath)
	defer os.Remove(STTInFilePath)
	defer os.Remove(STTOutFilePath)

	err := os.WriteFile(STTInFilePath, audioData.data, 0644)
	if err != nil {
		return "", fmt.Errorf("can't write stt input file: %w", err)
	}

	var args []string
	if reqParams.Language != "" {
		args = append(args, "--language", reqParams.Language)
	}
	args = append(args, STTInFilePath)
	cmd := NewCommand(ctx, params.STTBin, args...)
	cmd.Dir = path.Dir(params.STTBin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("STT error: %w: %s", err, string(output))
	}

	result, err := os.ReadFile(STTOutFilePath)
	if err != nil {
		return "", fmt.Errorf("can't read stt output file: %w", err)
	}
	return string(result), nil
}
