# MOTIVIC CONVERTOR
Future site of file I/O microservice to convert files from
1. MOTIF => MIDI
2. MIDI => WAV
3. WAV => MP3
4. MOTIF <==> JSON

## RUN
to test MIDI => WAV conversion
```bash
go build
./motivic_convertor -input test.midi -format wav -output test
# test generated aufio file
afplay test.wav
```