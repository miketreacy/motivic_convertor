package main

// leverage these packages for conversion:
// "github.com/go-audio/midi"
// "github.com/go-audio/wav"
import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/midi"
	"github.com/go-audio/wav"
)

// Future site of file I/O microservice to convert files from
// 1. MOTIF => MIDI
// 2. MIDI => WAV
// 3. WAV => MP3
// 4. MOTIF <==> JSON
var (
	flagInput  = flag.String("input", "", "The file to convert")
	flagFormat = flag.String("format", "wav", "The format to convert to (wav or aiff)")
	flagOutput = flag.String("output", "out", "The output filename")
)

func main() {
	// set up CLI
	flag.Parse()
	if *flagInput == "" {
		fmt.Println("Provide an input file using the -input flag")
		os.Exit(1)
	}
	switch strings.ToLower(*flagFormat) {
	case "aiff", "aif":
		*flagFormat = "aiff"
	case "wave", "wav":
		*flagFormat = "wav"
	default:
		fmt.Println("Provide a valid -format flag")
		os.Exit(1)
	}
	f, err := os.Open(*flagInput)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var of *os.File
	outputFilename := fmt.Sprintf("%s.%s", *flagOutput, *flagFormat)

	// generate the sound file
	var outName string
	var format string
	switch strings.ToLower(*formatFlag) {
	case "aif", "aiff":
		format = "aif"
		outName = "generated.aiff"
	default:
		format = "wav"
		outName = "generated.wav"
	}

	o, err := os.Create(outName)
	if err != nil {
		panic(err)
	}
	defer o.Close()
	if err := encodeAudioFile(format, buf, o); err != nil {
		panic(err)
	}
	fmt.Println(outName, "generated")

}

// take a MIDI file buffer and return parsed music events (Motivic.Motif format)
func decodeMIDIFile(filePath string) ([]struct, error) {
	f, err := os.Open(*filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decodedFile := midi.NewDecoder(f)
	if err := p.Parse(); err != nil {
		return nil, err
	}
	parsedTracks := []struct
	for i, t := range decodedFile.Tracks {
		parsedTrack := parseMIDITrack(t)
		parsedTracks = append(parsedTracks, parsedTrack)
	}
}

func parseMIDITrack(track struct) struct {
	// serialize midi.Track to Motivic.Motif
	parsedEvents := []struct
	for i, e := range track.Events {
		parsedEvent := parseMIDIEvent(e)
		parsedEvents = append(parsedEvents, parsedEvent)
	}
	return struct 
}

func parseMIDIEvent(event struct) struct {
	// serialize midi.Event to Motivic.Note

}

// take motif and return slice of audio buffers
func motifAudioMap(motif []string) []audio.Buffer {
	// for note in motif generateAudioFrequency(note)

}

// take frequency, duration, bit depth, and sample rate and return audio buffer of one note
func generateAudioFrequency(freq float64, dur int, bitDepth int, sampleRate int) audio.Buffer {
	osc := generator.NewOsc(generator.WaveSine, float64(freq), fs)
	// our osc generates values from -1 to 1, we need to go back to PCM scale
	factor := float64(audio.IntMaxSignedValue(biteDepth))
	osc.Amplitude = factor
	data := make([]float64, fs**durationFlag)
	buf := &audio.FloatBuffer{Data: data, Format: audio.FormatMono44100}
	osc.Fill(buf)
	return buf
}

// take slice of audio buffers and write audio file
func encodeAudioFile(format string, buf audio.Buffer, w io.WriteSeeker) error {
	switch format {
	case "wav":
		e := wav.NewEncoder(w,
			buf.PCMFormat().SampleRate,
			16,
			buf.PCMFormat().NumChannels, 1)
		if err := e.Write(buf.AsIntBuffer()); err != nil {
			return err
		}
		return e.Close()
	case "aiff":
		e := aiff.NewEncoder(w,
			buf.PCMFormat().SampleRate,
			16,
			buf.PCMFormat().NumChannels)
		if err := e.Write(buf.AsIntBuffer()); err != nil {
			return err
		}
		return e.Close()
	default:
		return errors.New("unknown format")
	}

}
