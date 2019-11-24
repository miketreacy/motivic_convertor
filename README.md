# MOTIVIC CONVERTOR
Future site of file I/O microservice to convert files from
- [x] MIDI => MOTIF
- [ ] MOTIF => MIDI
- [x] MOTIF => WAV
- [ ] MOTIF => JSON
- [ ] WAV => MP3


## RUN
to test MIDI => WAV conversion
1. build app: `go build`
    1. to test CLI:
        1. `./motivic_convertor -input input/test.midi -format wav -output test`
        1. test generated WAV file: `afplay test.wav`
    1. to test web server:
        1. `./motivic_convertor`
        1. go to `localhost:8080`