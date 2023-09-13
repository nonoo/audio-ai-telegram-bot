import torchaudio
from audiocraft.models import MusicGen
from audiocraft.data.audio import audio_write
import argparse
import sys

def arg_parse() -> tuple:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input_file", type=str, help="input file")
    parser.add_argument("--description", type=str, help="description")
    parser.add_argument("--duration", type=int, default=8, help="duration in seconds")
    parser.add_argument("--output_path", type=str, help="output path")

    args = parser.parse_args()
    sys.argv = sys.argv[:1]

    return args

def main():
    args = arg_parse()

    model = MusicGen.get_pretrained('facebook/musicgen-melody')
    model.set_generation_params(duration=args.duration)
    wav = model.generate_unconditional(4)    # generates 4 unconditional audio samples
    descriptions = [args.description]
    wav = model.generate(descriptions)

    melody, sr = torchaudio.load(args.input_file)
    # generates using the melody from the given audio and the provided descriptions.
    wav = model.generate_with_chroma(descriptions, melody[None].expand(1, -1, -1), sr)

    for idx, one_wav in enumerate(wav):
        # Will save under {idx}.wav, with loudness normalization at -14 db LUFS.
        audio_write(f'{args.output_path}/{idx}', one_wav.cpu(), model.sample_rate, strategy="loudness", loudness_compressor=True)

if __name__ == "__main__":
    main()
