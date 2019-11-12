package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
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
	// populate Motivic config values in memory
	initMotivicConfig()
	for _, p := range config.Pitches {
		fmt.Printf("CONFIG PITCH:\t%+v\n", p)
	}

	// set up CLI IO
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

	// parse the MIDI file to Motivic format
	motifs, err := decodeMIDIFile(flagInput)
	if err != nil {
		fmt.Println("ERROR: decodeMIDIFile", err)
		panic(err)
	}
	// for now Motivic only supports monophonic melodies so just grab the first track
	motif := motifs[0]

	for _, n := range motif.Notes {
		fmt.Printf("MOTIF NOTE:\t%+v\n", n)
	}

	// convert Motif to audio buffers
	motifBuffers := motifAudioMap(motif)
	// generate the audio file
	outputFilename := fmt.Sprintf("%s.%s", *flagOutput, *flagFormat)

	o, err := os.Create(outputFilename)
	if err != nil {
		panic(err)
	}
	defer o.Close()
	if err := encodeAudioFile(*flagFormat, motifBuffers, o); err != nil {
		panic(err)
	}
	fmt.Println(*flagOutput, "generated")
}
