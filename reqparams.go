package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

type ReqParamsTTS struct {
	Model string
}

func (r ReqParamsTTS) String() string {
	return "🗣️ " + r.Model
}

type ReqParamsSTT struct {
	Language string
}

func (r ReqParamsSTT) String() string {
	lang := r.Language
	if lang == "" {
		lang = "Autodetect"
	}
	return "🏳️‍🌈 " + lang
}

type ReqParamsMDX struct {
	FullOutput bool
}

func (r ReqParamsMDX) String() string {
	if r.FullOutput {
		return "👑 Full output"
	}
	return ""
}

type ReqParamsRVC struct {
	Model           string
	Pitch           int
	PitchSet        bool
	Method          string
	FilterRadius    int
	FilterRadiusSet bool
	IndexRate       float64
	IndexRateSet    bool
	RMSMixRate      float64
	RMSMixRateSet   bool
}

func (r ReqParamsRVC) String() string {
	args := []string{"🤡 " + r.Model + " 🎹 Method: " + r.Method}
	if r.PitchSet {
		args = append(args, "Pitch: "+fmt.Sprint(r.Pitch))
	}
	if r.FilterRadiusSet {
		args = append(args, "Filter radius: "+fmt.Sprint(r.FilterRadius))
	}
	if r.IndexRateSet {
		args = append(args, "Index rate: "+fmt.Sprint(r.IndexRate))
	}
	if r.RMSMixRateSet {
		args = append(args, "RMS mix rate: "+fmt.Sprint(r.RMSMixRate))
	}
	return strings.Join(args, " ")
}

type ReqParamsMusicgen struct {
	LengthSec    int
	LengthSecSet bool
}

func (r ReqParamsMusicgen) String() string {
	if r.LengthSecSet {
		return "🎹 Length: " + fmt.Sprint(r.LengthSec) + "s"
	}
	return ""
}

type ReqParamsAudiogen struct {
	LengthSec    int
	LengthSecSet bool
}

func (r ReqParamsAudiogen) String() string {
	if r.LengthSecSet {
		return "🎹 Length: " + fmt.Sprint(r.LengthSec) + "s"
	}
	return ""
}

type ReqParams interface {
	String() string
}

// Returns -1 as firstCmdCharAt if no params have been found in the given string.
func ReqParamsParse(ctx context.Context, s string, reqParams ReqParams) (prompt string, err error) {
	lexer := shlex.NewLexer(strings.NewReader(s))

	var reqParamsTTS *ReqParamsTTS
	var reqParamsSTT *ReqParamsSTT
	var reqParamsMDX *ReqParamsMDX
	var reqParamsRVC *ReqParamsRVC
	var reqParamsMusicgen *ReqParamsMusicgen
	var reqParamsAudiogen *ReqParamsAudiogen
	switch v := reqParams.(type) {
	case *ReqParamsTTS:
		reqParamsTTS = v
	case *ReqParamsSTT:
		reqParamsSTT = v
	case *ReqParamsMDX:
		reqParamsMDX = v
	case *ReqParamsRVC:
		reqParamsRVC = v
	case *ReqParamsMusicgen:
		reqParamsMusicgen = v
	case *ReqParamsAudiogen:
		reqParamsAudiogen = v
	default:
		return "", fmt.Errorf("invalid reqParams type")
	}

	for {
		token, lexErr := lexer.Next()
		if lexErr != nil { // No more tokens?
			break
		}

		if token[0] != '-' {
			toks := []string{token}
			for {
				token, lexErr := lexer.Next()
				if lexErr != nil { // No more tokens?
					return strings.Join(toks, " "), nil
				}
				toks = append(toks, token)
			}
		}

		attr := strings.ToLower(token[1:])
		validAttr := false

		switch attr {
		case "model", "m":
			if reqParamsTTS == nil || reqParamsRVC == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			if reqParamsTTS != nil {
				reqParamsTTS.Model = val
			} else if reqParamsRVC != nil {
				reqParamsRVC.Model = val
			}
			fmt.Println("model:", val)
			validAttr = true
		case "lang":
			if reqParamsSTT == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			reqParamsSTT.Language = val
			validAttr = true
		case "full", "f":
			if reqParamsMDX == nil {
				break
			}
			reqParamsMDX.FullOutput = true
			validAttr = true
		case "pitch", "p":
			if reqParamsRVC == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			reqParamsRVC.Pitch, err = strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid pitch value")
			}
			reqParamsRVC.PitchSet = true
			validAttr = true
		case "method":
			if reqParamsRVC == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			reqParamsRVC.Method = val
			validAttr = true
		case "filter-radius":
			if reqParamsRVC == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			reqParamsRVC.FilterRadius, err = strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid filter radius value")
			}
			reqParamsRVC.FilterRadiusSet = true
			validAttr = true
		case "index-rate":
			if reqParamsRVC == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			reqParamsRVC.IndexRate, err = strconv.ParseFloat(val, 32)
			if err != nil {
				return "", fmt.Errorf("invalid index rate value")
			}
			reqParamsRVC.IndexRateSet = true
			validAttr = true
		case "rms-mix-rate":
			if reqParamsRVC == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			reqParamsRVC.RMSMixRate, err = strconv.ParseFloat(val, 32)
			if err != nil {
				return "", fmt.Errorf("invalid rms mix rate value")
			}
			reqParamsRVC.RMSMixRateSet = true
			validAttr = true
		case "length", "l":
			if reqParamsMusicgen == nil || reqParamsAudiogen == nil {
				break
			}
			val, lexErr := lexer.Next()
			if lexErr != nil {
				return "", fmt.Errorf(attr + " is missing value")
			}
			if reqParamsMusicgen != nil {
				reqParamsMusicgen.LengthSec, err = strconv.Atoi(val)
				reqParamsMusicgen.LengthSecSet = true
			} else if reqParamsAudiogen != nil {
				reqParamsMusicgen.LengthSec, err = strconv.Atoi(val)
				reqParamsMusicgen.LengthSecSet = true
			}
			if err != nil {
				return "", fmt.Errorf("invalid length value")
			}
			validAttr = true
		}

		if !validAttr {
			return "", fmt.Errorf("unknown param " + token)
		}
	}

	return
}
