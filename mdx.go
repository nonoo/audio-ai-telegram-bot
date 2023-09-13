package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
)

type MDX struct {
}

var MDXInFilePath = os.TempDir() + "/mdx.wav"
var MDXInstrumFilePath = os.TempDir() + "/mdx_instrum.wav"
var MDXInstrum2FilePath = os.TempDir() + "/mdx_instrum2.wav"
var MDXVocalsFilePath = os.TempDir() + "/mdx_vocals.wav"
var MDXBassFilePath = os.TempDir() + "/mdx_bass.wav"
var MDXDrumsFilePath = os.TempDir() + "/mdx_drums.wav"
var MDXOtherFilePath = os.TempDir() + "/mdx_other.wav"

func (m *MDX) CleanupOutputFiles() {
	os.Remove(MDXInstrumFilePath)
	os.Remove(MDXInstrum2FilePath)
	os.Remove(MDXVocalsFilePath)
	os.Remove(MDXBassFilePath)
	os.Remove(MDXDrumsFilePath)
	os.Remove(MDXOtherFilePath)
}

func (m *MDX) MDX(ctx context.Context, reqParams ReqParamsMDX, audioData AudioFileData) ([]UploadFileData, error) {
	os.Remove(MDXInFilePath)
	defer os.Remove(MDXInFilePath)
	m.CleanupOutputFiles()

	err := os.WriteFile(MDXInFilePath, audioData.data, 0644)
	if err != nil {
		return nil, fmt.Errorf("can't write mdx input file: %w", err)
	}

	var args []string
	if !reqParams.FullOutput {
		args = append(args, "--vocals_only", "True")
	}
	args = append(args, "--input_audio", MDXInFilePath, "--output_folder", os.TempDir())
	cmd := NewCommand(ctx, params.MDXBin, args...)
	cmd.Dir = path.Dir(params.MDXBin)

	var lineBeforePercent string
	var percent int
	canceled, err := cmd.RunAndProcessOutput(func(line string) {
		re := regexp.MustCompile(`(\d+)%\|`)
		match := re.FindStringSubmatch(line)
		if len(match) > 1 && match[1] != "" {
			percent, err = strconv.Atoi(match[1])
			if err == nil {
				fmt.Print("    progress: ", lineBeforePercent, " ", percent, "%\n")
				reqQueue.currentEntry.entry.sendProcessUpdate(ctx, lineBeforePercent, percent)
			}
		} else {
			if line != "" {
				lineBeforePercent = line
				reqQueue.currentEntry.entry.sendProcessUpdate(ctx, lineBeforePercent, percent)
			}
		}
	})

	if canceled {
		m.CleanupOutputFiles()
		return nil, nil
	}

	if err != nil {
		m.CleanupOutputFiles()
		return nil, fmt.Errorf("MDX run error: %w", err)
	}

	var result []UploadFileData
	if _, err := os.Stat(MDXInstrumFilePath); !os.IsNotExist(err) {
		r, err := converter.ConvertToMP3(ctx, MDXInstrumFilePath)
		if err != nil {
			m.CleanupOutputFiles()
			return nil, err
		}
		result = append(result, UploadFileData{
			r:        r,
			filename: fileNameWithoutExt(audioData.filename) + " (Instrumental).mp3",
		})
	}
	if _, err := os.Stat(MDXInstrum2FilePath); !os.IsNotExist(err) {
		r, err := converter.ConvertToMP3(ctx, MDXInstrum2FilePath)
		if err != nil {
			m.CleanupOutputFiles()
			return nil, err
		}
		result = append(result, UploadFileData{
			r:        r,
			filename: fileNameWithoutExt(audioData.filename) + " (Instrumental2).mp3",
		})
	}
	if _, err := os.Stat(MDXVocalsFilePath); !os.IsNotExist(err) {
		r, err := converter.ConvertToMP3(ctx, MDXVocalsFilePath)
		if err != nil {
			m.CleanupOutputFiles()
			return nil, err
		}
		result = append(result, UploadFileData{
			r:        r,
			filename: fileNameWithoutExt(audioData.filename) + " (Vocals).mp3",
		})
	}
	if _, err := os.Stat(MDXBassFilePath); !os.IsNotExist(err) {
		r, err := converter.ConvertToMP3(ctx, MDXBassFilePath)
		if err != nil {
			m.CleanupOutputFiles()
			return nil, err
		}
		result = append(result, UploadFileData{
			r:        r,
			filename: fileNameWithoutExt(audioData.filename) + " (Bass).mp3",
		})
	}
	if _, err := os.Stat(MDXDrumsFilePath); !os.IsNotExist(err) {
		r, err := converter.ConvertToMP3(ctx, MDXDrumsFilePath)
		if err != nil {
			m.CleanupOutputFiles()
			return nil, err
		}
		result = append(result, UploadFileData{
			r:        r,
			filename: fileNameWithoutExt(audioData.filename) + " (Drums).mp3",
		})
	}
	if _, err := os.Stat(MDXOtherFilePath); !os.IsNotExist(err) {
		r, err := converter.ConvertToMP3(ctx, MDXOtherFilePath)
		if err != nil {
			m.CleanupOutputFiles()
			return nil, err
		}
		result = append(result, UploadFileData{
			r:        r,
			filename: fileNameWithoutExt(audioData.filename) + " (Other).mp3",
		})
	}

	return result, nil
}
