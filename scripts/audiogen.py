import torchaudio
from audiocraft.models import AudioGen
from audiocraft.data.audio import audio_write
import argparse
import sys

def arg_parse() -> tuple:
    parser = argparse.ArgumentParser()
    parser.add_argument("--description", type=str, help="description")
    parser.add_argument("--duration", type=int, default=8, help="duration in seconds")
    parser.add_argument("--output_path", type=str, help="output path")

    args = parser.parse_args()
    sys.argv = sys.argv[:1]

    return args

def main():
    args = arg_parse()

    model = AudioGen.get_pretrained('facebook/audiogen-medium')
    model.set_generation_params(duration=args.duration)
    descriptions = [args.description]
    wav = model.generate(descriptions)

    for idx, one_wav in enumerate(wav):
        # Will save under {idx}.wav, with loudness normalization at -14 db LUFS.
        audio_write(f'{args.output_path}/{idx}', one_wav.cpu(), model.sample_rate, strategy="loudness", loudness_compressor=True)

if __name__ == "__main__":
    main()
