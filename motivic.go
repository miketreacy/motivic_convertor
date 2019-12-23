package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Pitch : scientific notation pitch
type Pitch struct {
	Name   string `json:"name"`
	Octave int    `json:"octave"`
	Value  int    `json:"value"`
}

// Note : Motivic.Note class
type Note struct {
	Value    int `json:"value"` // scientific notation
	Duration int `json:"duration"`

	// TODO: migrate to computed property methods
	// computed
	Name   string `json:"name"`   // Scientific pitch notation note https://en.wikipedia.org/wiki/Scientific_pitch_notation
	Octave int    `json:"octave"` // Scientific pitch notation octave https://en.wikipedia.org/wiki/Scientific_pitch_notation
	Pitch  string `json:"pitch"`  // Scientific pitch notation (note + octave) https://en.wikipedia.org/wiki/Scientific_pitch_notation
}

// Note factory function
func newNote(v int, d int) Note {
	name, octave := getNoteNameAndOctave(v)
	pitchStr := fmt.Sprintf("%v%d", name, octave)
	n := Note{Value: v, Duration: d, Name: name, Octave: octave, Pitch: pitchStr}
	return n
}

// MotifNote : Motivic.Note decorated with motif-relative computed fields
type MotifNote struct {
	Note
	// relative (to Motif)
	// TODO: migrate to computed property methods
	Steps        int `json:"steps"`        // relative to Motif.Notes[0].Value
	StartingBeat int `json:"startingBeat"` // relative to Motif.Notes[0].StartingBeat
	Interval     int `json:"interval"`     // relative to Motif.Key
}

// Tempo : Motivic.Tempo class
type Tempo struct {
	Type  string `json:"type"`
	Units int    `json:"units"`
}

// TimeSignature : Motivic.TimeSignature class
type TimeSignature struct {
	Beat int `json:"beat"`
	Unit int `json:"unit"`
}

// Motif : Motivic.Motif melody class
type Motif struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
	Mode string `json:"mode"`
	Tempo
	TimeSignature
	Notes []MotifNote
}

// MotivicConfig : Motivic music theory config
type MotivicConfig struct {
	Frequencies [][]float64 `json:"frequencies"`
	Notes       []string    `json:"notes"`
	Pitches     []Pitch
}

var config MotivicConfig

// Index : simple utils for getting index of a slice element
func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func getNoteNameAndOctave(value int) (string, int) {
	// notes with negative value are rests
	if value < 0 {
		return "", -1
	}
	note := config.Pitches[value-1]
	return note.Name, note.Octave
}

func getPitchFrequency(pitch string, octave int) float64 {
	// handle rests - where pitch and octave are falsey
	if pitch == "" {
		return 0.00
	}
	idx := Index(config.Notes, pitch)
	return config.Frequencies[octave][idx]
}
func (c *MotivicConfig) setPitches() {
	var pitches []Pitch
	for octIdx := range c.Frequencies {
		for noteIdx, n := range c.Notes {
			p := Pitch{Name: n, Octave: octIdx, Value: (octIdx * 12) + noteIdx + 1}
			pitches = append(pitches, p)
		}
	}
	c.Pitches = pitches
}

func initMotivicConfig() {
	// read in app config from json file in root
	f, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	var c MotivicConfig
	json.Unmarshal(f, &c)
	c.setPitches()

	// assign config values to global var
	config = c
}

func getDurationInSeconds(dur int, t Tempo, ts TimeSignature) float64 {
	beatsPerSec := float64(t.Units) / float64(60)
	secsPerBeat := float64(1) / float64(beatsPerSec)
	beatsPerNote := float64(dur) / float64(ts.Beat*ts.Unit)
	durSecs := secsPerBeat * beatsPerNote
	return float64(durSecs)
}
