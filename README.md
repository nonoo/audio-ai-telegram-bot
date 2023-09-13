# Audio AI Telegram Bot

This is a Telegram Bot frontend for processing audio with several AI tools:

- [Coqui AI](https://github.com/coqui-ai/TTS) for text to speech
- [Whisper](https://github.com/openai/whisper) for speech to text
- [MDX23v2](https://github.com/ZFTurbo/MVSEP-MDX23-music-separation-model) for music separation
- [RVC WebUI](https://github.com/RVC-Project/Retrieval-based-Voice-Conversion-WebUI) for retrieval based voice conversion
- [AudioCraft](https://github.com/facebookresearch/audiocraft) for generating music and audio

<p align="center"><img src="demo.gif?raw=true"/></p>

The bot displays the progress (if available) and further information during
processing by responding to the message with the prompt. Requests are queued,
only one gets processed at a time.

The bot uses the
[Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api).
Rendered data are not saved on disk. Tested on Linux, but should be able
to run on other operating systems.

## Compiling

You'll need Go installed on your computer. Install a recent package of `golang`.
Then:

```
go get github.com/nonoo/audio-ai-telegram-bot
go install github.com/nonoo/audio-ai-telegram-bot
```

This will typically install `audio-ai-telegram-bot` into `$HOME/go/bin`.

Or just enter `go build` in the cloned Git source repo directory.

## Prerequisites

Create a Telegram bot using [BotFather](https://t.me/BotFather) and get the
bot's `token`.

### Coqui AI

- Follow the [installations steps](https://github.com/coqui-ai/TTS) and make sure
  the `tts` command is available.

- Create a shell script in the Coqui AI directory with the following contents:

```bash
#!/bin/bash
tts "$@" --text "`cat`"
```

- Set this shell script as the TTS binary for the bot using the `-tts-bin` command
  line argument.

### Whisper

- Follor the [installation steps](https://github.com/openai/whisper) and make sure
  the `whisper` command is available.
- Create a shell script in the Whisper directory with the following contents:

```bash
#!/bin/bash
env/bin/whisper --model large-v2 --model_dir . --output_format txt --output_dir /tmp "$@"
```

- Set this shell script as the STT binary for the bot using the `-stt-bin` command
  line argument.

### MDX23v2

- Clone the [MDX23v2](https://github.com/ZFTurbo/MVSEP-MDX23-music-separation-model) repo
- Enter into the cloned directory
- `python3 -m venv env`
- `pip install -r requirements.txt`
- Create a shell script in the repo directory with the following contents:

```bash
#!/bin/bash
. env/bin/activate
python inference.py $*
```

- Set this shell script as the MDX binary for the bot using the `-mdx-bin` command
  line argument.

### RVC WebUI

- Clone the [RVC WebUI](https://github.com/RVC-Project/Retrieval-based-Voice-Conversion-WebUI) repo
- Enter into the cloned directory
- `python3 -m venv env`
- `pip install -r requirements.txt`
- Create a shell script in the repo directory with the following contents:

```bash
#!/bin/bash
. env/bin/activate
python tools/infer_cli.py $*
```

- Set this shell script as the RVC binary for the bot using the `-rvc-bin` command
  line argument.
- Set the RVC model path directory using the `-rvc-model-path` command line argument.
  This is usually located at `rvc/assets/weights`

### AudioCraft

- Follow the [installation steps]([AudioCraft](https://github.com/facebookresearch/audiocraft))

#### Musicgen

- Set the `scripts/musicgen.sh` shell script as the Musicgen binary for the bot using
  the `-musicgen-bin` command line argument.

#### Audiogen

- Set the `scripts/audiogen.sh` shell script as the Audiogen binary for the bot using
  the `-audiogen-bin` command line argument.

## Running

You can get the available command line arguments with `-h`.
Mandatory arguments are:

- `-bot-token`: set this to your Telegram bot's `token`
-	`-tts-bin`: path of the TTS binary
	`-stt-bin`: path to the STT binary
	`-mdx-bin`: path to the MDX binary
	`-rvc-bin`: path to the RVC binary
	`-rvc-model-path`: path to the RVC weights directory
	`-musicgen-bin`: path to the Musicgen binary
	`-audiogen-bin`: path to the Audiogen binary

Set your Telegram user ID as an admin with the `-admin-user-ids` argument.
Admins will get a message when the bot starts.

Other user/group IDs can be set with the `-allowed-user-ids` and
`-allowed-group-ids` arguments. IDs should be separated by commas.

You can get Telegram user IDs by writing a message to the bot and checking
the app's log, as it logs all incoming messages.

All command line arguments can be set through OS environment variables.
Note that using a command line argument overwrites a setting by the environment
variable. Available OS environment variables are:

- `BOT_TOKEN`
- `ALLOWED_USERIDS`
- `ADMIN_USERIDS`
- `ALLOWED_GROUPIDS`
- `TTS_BIN`
- `TTS_DEFAULT_MODEL`
- `STT_BIN`
- `MDX_BIN`
- `RVC_BIN`
- `RVC_MODEL_PATH`
- `RVC_DEFAULT_MODEL`
- `MUSICGEN_BIN`
- `AUDIOGEN_BIN`

## Supported commands

- `/aaitts` (-m [model]) [prompt] - text to speech
- `/aaitts-models` - list text to speech models
- `/aaistt` (-lang [language]) - speech to text
- `/aaimdx` (-f) - music and voice separation (-f enables full output including instrument and bassline tracks)
- `/aairvc` (model) (-m [model]) (-p [pitch]) (-method [method]) (-filter-radius [v]) (-index-rate [v]) (-rms-mix-rate [v]) - retrieval based voice conversion
- `/aairvc-models` - list rvc models
- `/aaimusicgen` (-l [sec]) [prompt] - generate music based on given audio file and prompt
- `/aaiaudiogen` (-l [sec]) [prompt] - generate audio
- `/aaicancel` - cancel current req
- `/aaihelp` - show this help

You can also use the `!` command character instead of `/`.

You don't need to enter the `/aaitts` command if you send a prompt to the bot using
a private chat.

## Donations

If you find this bot useful then [buy me a beer](https://paypal.me/ha2non). :)
