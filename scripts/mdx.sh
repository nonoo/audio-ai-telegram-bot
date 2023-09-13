#!/bin/bash
. env/bin/activate
python -u inference.py $*
