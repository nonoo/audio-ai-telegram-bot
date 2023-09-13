package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-telegram/bot/models"
)

type RVC struct {
}

var RVCInFilePath = os.TempDir() + "/rvc-in.wav"
var RVCOutFilePath = os.TempDir() + "/rvc-out.wav"

func (t *RVC) GetModels() ([]string, error) {
	var models []string
	err := filepath.Walk(params.RVCModelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".pth") {
			models = append(models, fileNameWithoutExt(info.Name()))
		}
		return nil
	})
	return models, err
}

func (t *RVC) ListModels(ctx context.Context, msg *models.Message) {
	models, err := rvc.GetModels()
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't list models: "+err.Error())
		return
	}

	sendReplyToMessage(ctx, msg, "🤡 Available RVC models: "+strings.Join(models, ", "))
}

func (t *RVC) CleanupOutputFiles() {
	os.Remove(RVCOutFilePath)
}

func (t *RVC) RVC(ctx context.Context, reqParams ReqParamsRVC, audioData AudioFileData) (io.ReadCloser, error) {
	rvc.CleanupOutputFiles()

	defer os.Remove(RVCInFilePath)
	err := os.WriteFile(RVCInFilePath, audioData.data, 0644)
	if err != nil {
		return nil, fmt.Errorf("can't write rvc input file: %w", err)
	}

	modelFilename := reqParams.Model
	if !strings.HasSuffix(modelFilename, ".pth") {
		modelFilename += ".pth"
	}
	modelPath := path.Join(params.RVCModelPath, modelFilename)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model %s not found", reqParams.Model)
	}

	indexFilename := fileNameWithoutExt(modelFilename) + "_added.index"
	indexPath := path.Join(params.RVCModelPath, indexFilename)
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("index not found: %s", indexPath)
	}

	args := []string{"--input_path", RVCInFilePath, "--model_name", modelFilename,
		"--index_path", indexPath, "--opt_path", RVCOutFilePath, "--f0method", reqParams.Method}
	if reqParams.FilterRadiusSet {
		args = append(args, "--filter_radius", strconv.Itoa(reqParams.FilterRadius))
	}
	if reqParams.IndexRateSet {
		args = append(args, "--index_rate", fmt.Sprintf("%f", reqParams.IndexRate))
	}
	if reqParams.RMSMixRateSet {
		args = append(args, "--rms_mix_rate", fmt.Sprintf("%f", reqParams.RMSMixRate))
	}
	if reqParams.PitchSet {
		args = append(args, "--f0up_key", strconv.Itoa(reqParams.Pitch))
	}
	cmd := NewCommand(ctx, params.RVCBin, args...)
	cmd.Dir = path.Dir(params.RVCBin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		rvc.CleanupOutputFiles()
		return nil, fmt.Errorf("RVC error: %w: %s", err, string(output))
	}

	// Check output .wav file
	if stat, err := os.Stat(RVCOutFilePath); os.IsNotExist(err) || stat.Size() == 0 {
		rvc.CleanupOutputFiles()
		return nil, fmt.Errorf("output file not found: %s", RVCOutFilePath)
	}

	r, err := converter.ConvertToOpus(ctx, RVCOutFilePath)
	if err != nil {
		rvc.CleanupOutputFiles()
		return nil, err
	}

	return r, nil
}
