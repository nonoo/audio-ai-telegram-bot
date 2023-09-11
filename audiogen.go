package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

type Audiogen struct {
}

var AudiogenOutFilePath = os.TempDir() + "/0.wav"

func (a *Audiogen) CleanupOutputFiles() {
	os.Remove(AudiogenOutFilePath)
}

func (a *Audiogen) Audiogen(ctx context.Context, reqParams ReqParamsAudiogen, prompt string) (io.ReadCloser, error) {
	a.CleanupOutputFiles()

	args := []string{"--description", prompt, "--output_path", os.TempDir()}
	if reqParams.LengthSecSet {
		args = append(args, "--duration", strconv.Itoa(reqParams.LengthSec))
	}
	cmd := NewCommand(ctx, params.AudiogenBin, args...)
	cmd.Dir = path.Dir(params.AudiogenBin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		a.CleanupOutputFiles()
		return nil, fmt.Errorf("audiogen error: %w: %s", err, string(output))
	}

	// Check output .wav file
	if stat, err := os.Stat(AudiogenOutFilePath); os.IsNotExist(err) || stat.Size() == 0 {
		a.CleanupOutputFiles()
		return nil, fmt.Errorf("output file not found: %s", AudiogenOutFilePath)
	}

	r, err := converter.ConvertToOpus(ctx, AudiogenOutFilePath)
	if err != nil {
		a.CleanupOutputFiles()
		return nil, err
	}

	return r, nil
}
