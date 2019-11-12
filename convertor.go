package main

import (
	"errors"
	"fmt"
	"io"
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
	value := e.MIDINote - 11   //TODO: make sure conversion from MIDINote to MotifNote.value is correct!
	duration := e.Duration / 8 //TODO: conversion from ticks to MotifNote.duration is correct!
	n := newNote(value, duration)
	mn := MotifNote{
		Note:         n,
		StartingBeat: (e.Start / 8) + 1, //TODO: convert from ticks to MotifNote.startingBeat
	}
	return mn, nil
}

// take motif and return slice of audio buffers
func motifAudioMap(m Motif) []audio.FloatBuffer {
	var buffers []audio.FloatBuffer
	for _, n := range m.Notes {
		freq := getPitchFrequency(n.Name, n.Octave)
		fmt.Println("Note", n.Name, n.Octave, n.Pitch, "freq:", freq)
		buf := generateAudioFrequency(freq, n.Duration)
		buffers = append(buffers, *buf)
	}
	return buffers
}

// take frequency, duration, bit depth, and sample rate and return audio buffer of one note
func generateAudioFrequency(freq float64, dur int) *audio.FloatBuffer {
	// TODO: duration needs to be converted to seconds?
	// TODO: remove hardcoded bpm & time signature!!!
	ds := getDurationInSeconds(dur, 120, TimeSignature{4, 4})
	osc := generator.NewOsc(generator.WaveSine, float64(freq), audioSampleRate)
	// our osc generates values from -1 to 1, we need to go back to PCM scale
	factor := float64(audio.IntMaxSignedValue(audioBitDepth))
	osc.Amplitude = factor
	data := make([]float64, ds) //TODO: convert Motivic.duration to audio duration
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
