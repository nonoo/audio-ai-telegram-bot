package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

func (t *RVC) GetModelPaths(modelName string) (modelFilename, modelPath, indexPath string, err error) {
	modelFilename = modelName
	if !strings.HasSuffix(modelFilename, ".pth") {
		modelFilename += ".pth"
	}
	modelPath = path.Join(params.RVCModelPath, modelFilename)
	indexFilename := fileNameWithoutExt(modelFilename) + "_added.index"
	indexPath = path.Join(params.RVCModelPath, indexFilename)

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return modelFilename, modelPath, indexPath, fmt.Errorf("model %s not found", modelName)
	}

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return modelFilename, modelPath, indexPath, fmt.Errorf("index not found: %s", indexPath)
	}
	return
}

func (t *RVC) ModelExists(modelName string) bool {
	_, _, _, err := rvc.GetModelPaths(modelName)
	return err == nil
}

func (t *RVC) DeleteModel(modelName string) error {
	_, modelPath, indexPath, err := rvc.GetModelPaths(modelName)
	if err != nil {
		return err
	}

	if err := os.Remove(modelPath); err != nil {
		return err
	}
	if err := os.Remove(indexPath); err != nil {
		return err
	}
	return nil
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

	modelFilename, _, indexPath, err := rvc.GetModelPaths(reqParams.Model)
	if err != nil {
		return nil, err
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

func (t *RVC) TrainCleanupOutputFiles(modelName string) {
	rvcTrainBinPath := path.Dir(params.RVCTrainBin)
	path.Join(rvcTrainBinPath, "rvc-train-config.json")
	os.RemoveAll(path.Join(path.Dir(params.RVCTrainBin), "data", "training", "RVC", modelName))
}

func (t *RVC) Train(ctx context.Context, reqParams ReqParamsRVCTrain, audioData AudioFileData) error {
	modelFilename, modelPath, indexPath, err := rvc.GetModelPaths(reqParams.Model)
	if err == nil {
		return fmt.Errorf("model %s already exists", reqParams.Model)
	}

	rvc.TrainCleanupOutputFiles(reqParams.Model)

	trainDataDir, err := os.MkdirTemp("", "rvc-train")
	if err != nil {
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		return fmt.Errorf("can't create directory for training data: %w", err)
	}
	defer os.RemoveAll(trainDataDir)

	err = os.WriteFile(path.Join(trainDataDir, "in.wav"), audioData.data, 0644)
	if err != nil {
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		return fmt.Errorf("can't write rvc train input file: %w", err)
	}

	rvcTrainBinPath := path.Dir(params.RVCTrainBin)

	type TrainParams struct {
		Model     string `json:"model"`
		SrcDir    string `json:"src_dir"`
		Alg       string `json:"alg"`
		BatchSize int    `json:"batch_size"`
		Epochs    int    `json:"epochs"`
	}
	trainParams := TrainParams{
		Model:     reqParams.Model,
		SrcDir:    trainDataDir,
		Alg:       reqParams.Method,
		BatchSize: reqParams.BatchSize,
		Epochs:    reqParams.Epochs,
	}

	cfgFile, err := os.Create(path.Join(rvcTrainBinPath, "rvc-train-config.json"))
	if err != nil {
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		return fmt.Errorf("can't write rvc train config file: %w", err)
	}

	encoder := json.NewEncoder(cfgFile)
	err = encoder.Encode(trainParams)
	if err != nil {
		cfgFile.Close()
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		return fmt.Errorf("can't write rvc train config file: %w", err)
	}
	cfgFile.Close()

	cmd := NewCommand(ctx, params.RVCTrainBin)
	cmd.Dir = rvcTrainBinPath

	var prevPercent int
	canceled, err := cmd.RunAndProcessOutput(func(line string) {
		re := regexp.MustCompile(`^(\d+)\s+(\d+)\s+([\d.]+)`)
		match := re.FindStringSubmatch(line)

		if len(match) == 4 {
			if epochs, convErr := strconv.Atoi(match[1]); convErr == nil {
				percent := int(float32(epochs*100) / float32(reqParams.Epochs))
				var processDesc string
				if loss, convErr := strconv.ParseFloat(match[3], 32); convErr == nil {
					processDesc = processStr + " (loss: " + fmt.Sprintf("%.2f", loss) + ")"
				}
				if percent > prevPercent {
					fmt.Print("    progress: ", percent, "%\n")
					reqQueue.currentEntry.entry.sendProcessUpdate(ctx, processDesc, percent)
					prevPercent = percent
				}
			}
		}
	})

	if canceled {
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		_ = rvc.DeleteModel(reqParams.Model)
		return nil
	}

	if err != nil {
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		_ = rvc.DeleteModel(reqParams.Model)
		return fmt.Errorf("RVC train error: %w", err)
	}

	reqQueue.currentEntry.entry.sendProcessUpdate(ctx, "Copying results...", -1)

	dst := modelPath
	src := path.Join(path.Dir(params.RVCTrainBin), "data", "training", "RVC", reqParams.Model, "models", fmt.Sprint("e_", reqParams.Epochs-1), modelFilename)
	fmt.Println("  copying model file from", src, "to", dst)
	if err = copyFile(modelPath, src); err != nil {
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		_ = rvc.DeleteModel(reqParams.Model)
		return fmt.Errorf("can't copy model file to RVC model path: %w", err)
	}
	dst = indexPath
	src = path.Join(path.Dir(params.RVCTrainBin), "data", "training", "RVC", reqParams.Model, reqParams.Model+"_added.index")
	fmt.Println("  copying index file from", src, "to", dst)
	if err = copyFile(dst, src); err != nil {
		rvc.TrainCleanupOutputFiles(reqParams.Model)
		_ = rvc.DeleteModel(reqParams.Model)
		return fmt.Errorf("can't copy index file to RVC model path: %w", err)
	}

	return nil
}
