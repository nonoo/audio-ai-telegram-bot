package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

type Musicgen struct {
}

var MusicgenInFilePath = os.TempDir() + "/musicgen-in.wav"
var MusicgenOutFilePath = os.TempDir() + "/0.wav"

func (m *Musicgen) CleanupOutputFiles() {
	os.Remove(MusicgenOutFilePath)
}

func (m *Musicgen) Musicgen(ctx context.Context, reqParams ReqParamsMusicgen, prompt string, audioData AudioFileData) (io.ReadCloser, error) {
	m.CleanupOutputFiles()

	defer os.Remove(MusicgenInFilePath)
	err := os.WriteFile(MusicgenInFilePath, audioData.data, 0644)
	if err != nil {
		return nil, fmt.Errorf("can't write musicgen input file: %w", err)
	}

	args := []string{"--input_file", MusicgenInFilePath, "--description", prompt, "--output_path", os.TempDir()}
	if reqParams.LengthSecSet {
		args = append(args, "--duration", strconv.Itoa(reqParams.LengthSec))
	}
	cmd := NewCommand(ctx, params.MusicgenBin, args...)
	cmd.Dir = path.Dir(params.MusicgenBin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		m.CleanupOutputFiles()
		return nil, fmt.Errorf("musicgen error: %w: %s", err, string(output))
	}

	// Check output .wav file
	if stat, err := os.Stat(MusicgenOutFilePath); os.IsNotExist(err) || stat.Size() == 0 {
		m.CleanupOutputFiles()
		return nil, fmt.Errorf("output file not found: %s", MusicgenOutFilePath)
	}

	r, err := converter.ConvertToOpus(ctx, MusicgenOutFilePath)
	if err != nil {
		m.CleanupOutputFiles()
		return nil, err
	}

	return r, nil
}
