#!/bin/bash
env/bin/whisper --model large-v2 --model_dir . --output_format txt --output_dir /tmp "$@"
