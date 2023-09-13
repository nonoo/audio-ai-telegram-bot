#!/bin/bash
. env/bin/activate
python tools/infer_cli.py $*
