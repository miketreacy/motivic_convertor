# MOTIVIC CONVERTOR
Future site of file I/O microservice to convert files from
- [x] MIDI => MOTIF
- [ ] MOTIF => MIDI
- [x] MOTIF => WAV
- [ ] MOTIF => JSON
- [ ] WAV => MP3


## RUN
to test MIDI => WAV conversion
```bash
go build
./motivic_convertor -input test.midi -format wav -output test
# test generated aufio file
afplay test.wav
```