package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Note struct {
	Value    int `json:"value"` // scientific notation
	Duration int `json:"duration"`

	// TODO: migrate to computed property methods
	// computed
	Name   string `json:"name"`   // Scientific pitch notation note https://en.wikipedia.org/wiki/Scientific_pitch_notation
	Octave int    `json:"octave"` // Scientific pitch notation octave https://en.wikipedia.org/wiki/Scientific_pitch_notation
	Pitch  string `json:"pitch"`  // Scientific pitch notation (note + octave) https://en.wikipedia.org/wiki/Scientific_pitch_notation
}

type MotifNote struct {
	Note
	// relative (to Motif)
	// TODO: migrate to computed property methods
	Steps        int `json:"steps"`        // relative to Motif.Notes[0].Value
	StartingBeat int `json:"startingBeat"` // relative to Motif.Notes[0].StartingBeat
	Interval     int `json:"interval"`     // relative to Motif.Key
}

type Tempo struct {
	Type  string
	Units int
}

type TimeSignature struct {
	Beat int
	Unit int
}

type Motif struct {
	Id   string
	Name string
	Key  string
	Mode string
	Tempo
	TimeSignature
	Notes []Note
}

type MotivicConfig struct {
	Frequencies [][]float64 `json:"frequencies"`
	Notes       []string    `json:"notes"`
}

var config MotivicConfig

func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func getPitchFrequency(pitch string, octave int) float64 {
	// freqs := [][]float64{
	// 	{16.351, 17.324, 18.354, 19.445, 20.601, 21.827, 23.124, 24.499, 25.956, 27.5, 29.135, 30.868},
	// 	{32.703, 34.648, 36.708, 38.891, 41.203, 43.654, 46.249, 48.999, 51.913, 55, 58.27, 61.735},
	// 	{65.406, 69.296, 73.416, 77.782, 82.407, 87.307, 92.499, 97.999, 103.826, 110, 116.541, 123.471},
	// 	{130.813, 138.591, 146.832, 155.563, 164.814, 174.614, 184.997, 195.998, 207.652, 220, 233.082, 246.942},
	// 	{261.626, 277.183, 293.665, 311.127, 329.628, 349.228, 369.994, 391.995, 415.305, 440, 466.164, 493.883},
	// 	{523.251, 554.365, 587.33, 622.254, 659.255, 698.456, 739.989, 783.991, 830.609, 880, 932.328, 987.767},
	// 	{1046.502, 1108.731, 1174.659, 1244.508, 1318.51, 1396.913, 1479.978, 1567.982, 1661.219, 1760, 1864.655, 1975.533},
	// 	{2093.005, 2217.461, 2349.318, 2489.016, 2637.021, 2793.826, 2959.955, 3135.964, 3322.438, 3520, 3729.31, 3951.066},
	// 	{4186.009, 4434.922, 4698.636, 4978.032, 5274.042, 5587.652, 5919.91, 6271.928, 6644.876, 7040, 7458.62, 7902.132},
	// 	{8372.018, 8869.844, 9397.272, 9956.064, 10548.084, 11175.304, 11839.82, 12543.856, 13289.752, 14080, 14917.24, 15804.264},
	// }
	idx := Index(config.Notes, pitch)
	return config.Frequencies[octave][idx]
}

func initConfig() {
	// read in app config from json file in root
	f, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	var c MotivicConfig
	json.Unmarshal(f, &c)
	// assign config values to global var
	config = c
}
