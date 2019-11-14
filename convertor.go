package main

import (
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

const audioBitDepth int = 16
const audioSampleRate int = 44100

// take a MIDI file buffer and return parsed music events (Motivic.Motif format)
func decodeMIDIFile(filePath *string) ([]Motif, error) {
	f, err := os.Open(*filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decodedFile := midi.NewDecoder(f)
	if err := decodedFile.Parse(); err != nil {
		return nil, err
	}
	var parsedTracks []Motif
	for _, t := range decodedFile.Tracks {
		parsedTrack, err := parseMIDITrack(t)
		if err != nil {
			fmt.Println("ERROR parsing track", err)
			panic(err)
		}
		parsedTracks = append(parsedTracks, parsedTrack)
	}
	return parsedTracks, nil
}

func parseMIDITrack(track *midi.Track) (Motif, error) {
	// serialize midi.Track to Motivic.Motif
	var parsedEvents []MotifNote
	for _, e := range track.AbsoluteEvents() {
		parsedEvent, err := parseMIDIEvent(e)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		parsedEvents = append(parsedEvents, parsedEvent)
	}
	m := Motif{Notes: parsedEvents}
	return m, nil
}

func parseMIDIEvent(e *midi.AbsEv) (MotifNote, error) {
	// TODO: serialize midi.Event to Motivic.Note
	fmt.Printf("MIDI EVENT:\t%+v\n", e)
	// TODO: make sure conversion from MIDINote to MotifNote.value is correct!
	value := e.MIDINote - 11
	// TODO: conversion from ticks to MotifNote.duration is correct!
	// TODO: make sure that these are always both ints!
	duration := e.Duration / 8
	n := newNote(value, duration)
	mn := MotifNote{
		Note: n,
		// TODO: make sure this conversion from ticks to MotifNote.startingBeat is correct
		// TODO: make sure that these are always both ints!
		StartingBeat: (e.Start / 8) + 1,
	}
	return mn, nil
}

// take motif and return slice of audio buffers
func motifAudioMap(m Motif) []audio.FloatBuffer {
	// TODO: remove hardcoded bpm & time signature!!!
	bpm := 120
	ts := TimeSignature{4, 4}
	var buffers []audio.FloatBuffer
	for _, n := range m.Notes {
		freq := getPitchFrequency(n.Name, n.Octave)
		// TODO: duration needs to be converted to seconds?
		// TODO: fix this - right now am rounding up to nearest second
		ds := int(math.Ceil(getDurationInSeconds(n.Duration, bpm, ts)))
		fmt.Println("AUDIO BUFFER:", n.Name, n.Octave, n.Pitch, "freq:", freq, "secs:", ds)
		buf := generateAudioFrequency(freq, ds)
		buffers = append(buffers, *buf)
	}
	return buffers
}

// take frequency, duration, bit depth, and sample rate and return audio buffer of one note
func generateAudioFrequency(freq float64, durSecs int) *audio.FloatBuffer {
	osc := generator.NewOsc(generator.WaveSine, float64(freq), audioSampleRate)
	// our osc generates values from -1 to 1, we need to go back to PCM scale
	factor := float64(audio.IntMaxSignedValue(audioBitDepth))
	osc.Amplitude = factor
	// TODO: figure out significance of []float64 length here - currently length is note length rounded down to seconds
	data := make([]float64, durSecs)
	buf := &audio.FloatBuffer{Data: data, Format: audio.FormatMono44100}
	osc.Fill(buf)
	return buf
}

func encodeWAVFile(bufs []audio.FloatBuffer, w io.WriteSeeker) error {
	e := wav.NewEncoder(w, bufs[0].PCMFormat().SampleRate, 16, bufs[0].PCMFormat().NumChannels, 1)
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
		16,
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
