package main

import (
	"context"
	"fmt"
	"io"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type Converter struct {
}

func (c *Converter) ConvertToMP3(ctx context.Context, filePath string) (reader io.ReadCloser, err error) {
	reader, writer := io.Pipe()

	fmt.Print("  converting to mp3...\n")

	args := ffmpeg_go.KwArgs{"format": "mp3", "c:a": "mp3", "b:a": "320k"}
	ff := ffmpeg_go.Input(filePath).Output("pipe:1", args)
	ffCmd := ff.WithOutput(writer).Compile()

	// Creating a new cmd with a timeout context, which will kill the cmd if it takes too long.
	cmd := NewCommand(ctx, ffCmd.Args[0], ffCmd.Args[1:]...)
	cmd.Stdin = ffCmd.Stdin
	cmd.Stdout = ffCmd.Stdout

	// This goroutine handles copying from the input (either r or cmd.Stdout) to writer.
	go func() {
		err = cmd.Run()
		writer.Close()
	}()

	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("error converting to mp3: %w", err)
	}

	return reader, nil
}

func (c *Converter) ConvertToOpus(ctx context.Context, filePath string) (reader io.ReadCloser, err error) {
	reader, writer := io.Pipe()

	fmt.Print("  converting to opus...\n")

	args := ffmpeg_go.KwArgs{"format": "ogg", "c:a": "libopus", "b:a": "256k", "vbr": "on", "compression_level": "10"}
	ff := ffmpeg_go.Input(filePath).Output("pipe:1", args)
	ffCmd := ff.WithOutput(writer).Compile()

	// Creating a new cmd with a timeout context, which will kill the cmd if it takes too long.
	cmd := NewCommand(ctx, ffCmd.Args[0], ffCmd.Args[1:]...)
	cmd.Stdin = ffCmd.Stdin
	cmd.Stdout = ffCmd.Stdout

	// This goroutine handles copying from the input (either r or cmd.Stdout) to writer.
	go func() {
		err = cmd.Run()
		writer.Close()
	}()

	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("error converting to opus: %w", err)
	}

	return reader, nil
}
