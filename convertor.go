package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/generator"
	"github.com/go-audio/midi"
	"github.com/go-audio/wav"
)

// Setting global audio config for 16/44/mono
const audioBitDepth int = 16
const audioSampleRate int = 44100
const midiNoteValueOffset int = -11
const midiDurationValueDivisor int = 8
const defaultWaveForm generator.WaveType = generator.WaveSine
const wavFile string = "wav"

var audioFormat = audio.FormatMono44100
var waveForm = map[string]generator.WaveType{
	"sine":     generator.WaveSine,
	"triangle": generator.WaveTriangle,
	"square":   generator.WaveSqr,
	"saw":      generator.WaveSaw,
}

func convertMIDIFileToWAVFile(inputFileName string, outputFilePath string, wf string, c chan<- bool) {
	success := false
	// parse the MIDI file to Motivic format
	motifs, err := parseMIDIFile(inputFileName)
	if err != nil || len(motifs) == 0 {
		fmt.Println("ERROR: parseMIDIFile", err)
		c <- success
		return
	}
	// for now Motivic only supports monophonic melodies so just grab the first track
	if len(motifs) > 1 {
		fmt.Println("ERROR: MIDI file is not monophonic")
		c <- success
		return
	}
	motif := motifs[0]

	for _, n := range motif.Notes {
		fmt.Printf("MOTIF NOTE:\t%+v\n", n)
	}

	// convert Motif to audio buffers
	motifBuffers := motifAudioMap(motif, wf)
	// generate the audio file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("ERROR:", err)
		c <- success
		return
	}
	defer outputFile.Close()
	if err := encodeAudioFile(wavFile, motifBuffers, outputFile); err != nil {
		fmt.Println("ERROR: encodeAudioFile", err)
		c <- success
		return
	}
	fmt.Println("Audio file generated at", outputFilePath)
	c <- true
	return
}

// take a JSON file on disk and return parsed music events (Motivic.Motif format)
func parseJSONFile(filePath string) ([]Motif, error) {
	var parsedTracks []Motif
	var err error = nil
	defer func() {
		if panicErr := recover(); panicErr != nil {
			fmt.Println("Recovered in parseJSONFile", panicErr)
			parsedTracks = nil
			err = errors.New("JSON file failed to parse")
		}
	}()
	file, err := os.Open(filePath)
	if err != nil {
		return parsedTracks, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	parsedJSON := []map[string]interface{}{}

	// Read the array open bracket
	decoder.Token()

	data := map[string]interface{}{}
	for decoder.More() {
		decoder.Decode(&data)
		parsedJSON = append(parsedJSON, data)
	}
	// TODO: convert parsed JSON to Motif tracks
	return parsedTracks, err
}

// take a MIDI file on disk and return parsed music events (Motivic.Motif format)
func parseMIDIFile(filePath string) ([]Motif, error) {
	var parsedTracks []Motif
	var err error = nil
	defer func() {
		if panicErr := recover(); panicErr != nil {
			fmt.Println("Recovered in parseMIDIFile", panicErr)
			parsedTracks = nil
			err = errors.New("MIDI file failed to parse")
		}
	}()
	file, err := os.Open(filePath)
	if err != nil {
		return parsedTracks, err
	}
	defer file.Close()
	decodedFile := midi.NewDecoder(file)
	if err := decodedFile.Parse(); err != nil {
		return parsedTracks, err
	}

	for _, t := range decodedFile.Tracks {
		parsedTrack, err := parseMIDITrack(t)
		if err != nil {
			fmt.Println("ERROR parsing track", err)
			return parsedTracks, err
		}
		parsedTracks = append(parsedTracks, parsedTrack)
	}
	return parsedTracks, err
}

func parseMIDITrack(track *midi.Track) (Motif, error) {
	// serialize midi.Track to Motivic.Motif
	// TODO: remove hardcoded time signature - parse from MIDI file
	fmt.Printf("\n*midi.Track: \t%+v\n\n", track)
	printReflectionInfo(track)
	m := Motif{}
	if track == nil {
		return m, fmt.Errorf("ERROR: parseMIDITrack() - track is nil")
	}
	bpm := track.Events[0].Bpm
	t := Tempo{Type: "bpm", Units: int(bpm)}
	ts := TimeSignature{4, 4}
	var parsedEvents []MotifNote
	for _, e := range track.AbsoluteEvents() {
		parsedEvent, err := parseMIDIEvent(e)
		if err != nil {
			fmt.Println(err)
			return m, err
		}
		parsedEvents = append(parsedEvents, parsedEvent)
	}
	parsedEvents = getNotesWithInsertedRests(parsedEvents)
	m = Motif{Notes: parsedEvents, Tempo: t, TimeSignature: ts}
	return m, nil
}

func getNotesWithInsertedRests(events []MotifNote) []MotifNote {
	// MIDI doesn't treat rests as events so
	// fabricate rest notes to fill in the gaps in parsedEvents
	var notes []MotifNote
	for i, e := range events {
		beatPosition := 1
		if i != 0 {
			// this is the first event
			prevNote := events[i-1]
			beatPosition = prevNote.StartingBeat + prevNote.Duration
		}
		// this is not the first event
		if e.StartingBeat != beatPosition {
			// there is a gap where a rest should go
			// for now, Note with negative value is a Rest
			rest := newNote(-1, e.StartingBeat-beatPosition)
			mn := MotifNote{
				Note:         rest,
				StartingBeat: beatPosition,
			}
			// Insert the rest before this note
			notes = append(notes, mn)
		}
		// there is no gap so add the note
		notes = append(notes, e)
	}
	return notes
}

// converts MIDI event.MIDINote to Motivic.Note.value
func convertMIDINote(note int) int {
	return note + midiNoteValueOffset
}

func convertMIDINoteDuration(dur int) int {
	return dur / midiDurationValueDivisor
}

func parseMIDIEvent(e *midi.AbsEv) (MotifNote, error) {
	// TODO: serialize midi.Event to Motivic.Note
	fmt.Printf("MIDI EVENT:\t%+v\n", e)
	// TODO: make sure conversion from MIDINote to MotifNote.value is correct!
	// TODO: handle RESTS!!!
	value := convertMIDINote(e.MIDINote)
	// TODO: conversion from ticks to MotifNote.duration is correct!
	// TODO: make sure that these are always both ints!
	duration := convertMIDINoteDuration(e.Duration)
	n := newNote(value, duration)
	mn := MotifNote{
		Note: n,
		// TODO: make sure this conversion from ticks to MotifNote.startingBeat is correct
		// TODO: make sure that these are always both ints!
		StartingBeat: convertMIDINoteDuration(e.Start) + 1,
	}
	return mn, nil
}

// take motif and return slice of audio buffers
func motifAudioMap(m Motif, voice string) []audio.FloatBuffer {
	var buffers []audio.FloatBuffer
	for _, n := range m.Notes {
		freq := getPitchFrequency(n.Name, n.Octave)
		// TODO: duration needs to be converted to seconds?
		// TODO: fix this - right now am rounding up to nearest second
		ds := getDurationInSeconds(n.Duration, m.Tempo, m.TimeSignature)
		fmt.Println("AUDIO NOTE DATA:", n.Name, n.Octave, n.Pitch, "freq:", freq, "secs:", ds)
		// TODO: handle rests!!!
		buf := generateAudioFrequency(freq, ds, voice)
		buffers = append(buffers, *buf)
	}
	return buffers
}

// take motif and return slice of MIDI buffers
func motifMIDIMap(m Motif) {
	// TODO: write this
}

// take motif and return JSON representation
func motifJSONMap(m Motif) {
	// TODO: write this
}

// take frequency, duration, bit depth, and sample rate and return audio buffer of one note
func generateAudioFrequency(freq float64, durSecs float64, voice string) *audio.FloatBuffer {
	wf := waveForm[voice]
	if wf == 0 {
		wf = defaultWaveForm
	}
	osc := generator.NewOsc(wf, float64(freq), audioSampleRate)
	// our osc generates values from -1 to 1, we need to go back to PCM scale
	factor := float64(audio.IntMaxSignedValue(audioBitDepth))
	osc.Amplitude = factor
	// buf.Data slice has length bitDepth * seconds
	data := make([]float64, int(math.Ceil(float64(audioSampleRate)*durSecs)))
	buf := &audio.FloatBuffer{Data: data, Format: audioFormat}
	osc.Fill(buf)
	return buf
}

func encodeWAVFile(bufs []audio.FloatBuffer, w io.WriteSeeker) error {
	// APPROACH: iterate through buffers and encode each one sequentially
	e := wav.NewEncoder(w, bufs[0].PCMFormat().SampleRate, audioBitDepth, bufs[0].PCMFormat().NumChannels, 1)
	for _, b := range bufs {
		err := e.Write(b.AsIntBuffer())
		if err != nil {
			return err
		}
	}
	return e.Close()
}

func encodeAIFFile(bufs []audio.FloatBuffer, w io.WriteSeeker) error {
	e := aiff.NewEncoder(w,
		bufs[0].PCMFormat().SampleRate,
		audioBitDepth,
		bufs[0].PCMFormat().NumChannels)
	for _, b := range bufs {
		err := e.Write(b.AsIntBuffer())
		if err != nil {
			return err
		}
	}
	return e.Close()
}

// take slice of audio buffers and write audio file
func encodeAudioFile(format string, bufs []audio.FloatBuffer, w io.WriteSeeker) error {
	switch format {
	case "wav":
		return encodeWAVFile(bufs, w)
	case "aiff":
		return encodeAIFFile(bufs, w)
	default:
		return errors.New("unknown format")
	}
}

func encodeMIDIFile(bufs []audio.FloatBuffer, w io.WriteSeeker) {
	// TODO: write this
}

func encodeJSONFile(jsonData []byte, filePath string) {
	// TODO: write this
	file, _ := os.OpenFile(filePath, os.O_CREATE, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(jsonData)
}
