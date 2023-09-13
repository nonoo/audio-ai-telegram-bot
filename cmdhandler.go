package main

import (
	"context"

	"github.com/go-telegram/bot/models"
)

type cmdHandlerType struct{}

func (c *cmdHandlerType) TTS(ctx context.Context, prompt string, msg *models.Message) {
	reqParams := ReqParamsTTS{
		Model: params.TTSDefaultModel,
	}
	var err error
	prompt, err = ReqParamsParse(ctx, prompt, &reqParams)
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't parse params: "+err.Error())
		return
	}

	if prompt == "" {
		sendReplyToMessage(ctx, msg, errorStr+": empty prompt")
		return
	}
	if reqParams.Model == "" {
		sendReplyToMessage(ctx, msg, errorStr+": no model given")
		return
	}

	req := ReqQueueReq{
		Type:    ReqTypeTTS,
		Message: msg,
		Prompt:  prompt,
		Params:  reqParams,
	}
	reqQueue.Add(req)
}

func (c *cmdHandlerType) STT(ctx context.Context, msg *models.Message) {
	reqParams := ReqParamsSTT{}
	_, err := ReqParamsParse(ctx, msg.Text, &reqParams)
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't parse params: "+err.Error())
		return
	}

	req := ReqQueueReq{
		Type:    ReqTypeSTT,
		Message: msg,
		Params:  reqParams,
	}
	reqQueue.Add(req)
}

func (c *cmdHandlerType) MDX(ctx context.Context, msg *models.Message) {
	reqParams := ReqParamsMDX{}
	_, err := ReqParamsParse(ctx, msg.Text, &reqParams)
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't parse params: "+err.Error())
		return
	}

	req := ReqQueueReq{
		Type:    ReqTypeMDX,
		Message: msg,
		Params:  reqParams,
	}
	reqQueue.Add(req)
}

func (c *cmdHandlerType) RVC(ctx context.Context, msg *models.Message) {
	reqParams := ReqParamsRVC{
		Method:       "harvest",
		Model:        msg.Text,
		FilterRadius: 3,
	}

	if reqParams.Model == "" {
		reqParams.Model = params.RVCDefaultModel
	}

	_, err := ReqParamsParse(ctx, msg.Text, &reqParams)
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't parse params: "+err.Error())
		return
	}

	if reqParams.Model == "" {
		sendReplyToMessage(ctx, msg, errorStr+": no model given")
		return
	}

	req := ReqQueueReq{
		Type:    ReqTypeRVC,
		Message: msg,
		Params:  reqParams,
	}
	reqQueue.Add(req)
}

func (c *cmdHandlerType) RVCTrain(ctx context.Context, msg *models.Message) {
	reqParams := ReqParamsRVCTrain{
		Method:    "harvest",
		Model:     msg.Text,
		BatchSize: params.RVCTrainDefaultBatchSize,
		Epochs:    params.RVCTrainDefaultEpochs,
	}

	if reqParams.Model == "" {
		reqParams.Model = params.RVCDefaultModel
	}

	_, err := ReqParamsParse(ctx, msg.Text, &reqParams)
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't parse params: "+err.Error())
		return
	}

	if reqParams.Model == "" {
		sendReplyToMessage(ctx, msg, errorStr+": no model given")
		return
	}

	req := ReqQueueReq{
		Type:    ReqTypeRVC,
		Message: msg,
		Params:  reqParams,
	}
	reqQueue.Add(req)
}

func (c *cmdHandlerType) Musicgen(ctx context.Context, prompt string, msg *models.Message) {
	reqParams := ReqParamsMusicgen{}
	var err error
	prompt, err = ReqParamsParse(ctx, prompt, &reqParams)
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't parse params: "+err.Error())
		return
	}

	if prompt == "" {
		sendReplyToMessage(ctx, msg, errorStr+": empty prompt")
		return
	}

	req := ReqQueueReq{
		Type:    ReqTypeMusicgen,
		Message: msg,
		Prompt:  prompt,
		Params:  reqParams,
	}
	reqQueue.Add(req)
}

func (c *cmdHandlerType) Audiogen(ctx context.Context, prompt string, msg *models.Message) {
	reqParams := ReqParamsAudiogen{}
	var err error
	prompt, err = ReqParamsParse(ctx, prompt, &reqParams)
	if err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": can't parse params: "+err.Error())
		return
	}

	if prompt == "" {
		sendReplyToMessage(ctx, msg, errorStr+": empty prompt")
		return
	}

	req := ReqQueueReq{
		Type:    ReqTypeAudiogen,
		Message: msg,
		Prompt:  prompt,
		Params:  reqParams,
	}
	reqQueue.Add(req)
}

func (c *cmdHandlerType) Cancel(ctx context.Context, msg *models.Message) {
	if err := reqQueue.CancelCurrentEntry(ctx); err != nil {
		sendReplyToMessage(ctx, msg, errorStr+": "+err.Error())
	}
}

func (c *cmdHandlerType) Help(ctx context.Context, msg *models.Message, cmdChar string) {
	sendReplyToMessage(ctx, msg, "🤖 Audio AI Telegram Bot\n\n"+
		"Available commands:\n\n"+
		cmdChar+"aaitts (-m [model]) [prompt] - text to speech\n"+
		cmdChar+"aaitts-models - list text to speech models\n"+
		cmdChar+"aaistt (-lang [language]) - speech to text\n"+
		cmdChar+"aaimdx (-f) - music and voice separation (-f enables full output including instrument and bassline tracks)\n"+
		cmdChar+"aairvc (model) (-m [model]) (-p [pitch]) (-method [method]) (-filter-radius [v]) (-index-rate [v]) (-rms-mix-rate [v]) - retrieval based voice conversion\n"+
		cmdChar+"aairvc-train (model) (-m [model]) (-method [method]) (-batch-size [v]) (-epochs [v]) (-delete) - retrieval based voice conversion training\n"+
		cmdChar+"aairvc-models - list rvc models\n"+
		cmdChar+"aaimusicgen (-l [sec]) [prompt] - generate music based on given audio file and prompt\n"+
		cmdChar+"aaiaudiogen (-l [sec]) [prompt] - generate audio\n"+
		cmdChar+"aaicancel - cancel current req\n"+
		cmdChar+"aaihelp - show this help\n\n"+
		"For more information see https://github.com/nonoo/audio-ai-telegram-bot")
}
