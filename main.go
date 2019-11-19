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
	flagInput    = flag.String("input", "", "The file to convert")
	flagFormat   = flag.String("format", "wav", "The format to convert to (wav or aiff)")
	flagOutput   = flag.String("output", "out", "The output filename")
	flagWaveForm = flag.String("waveform", "sine", "The oscillator waveform to use")
)

func cleanUp() {
	// TODO: delete own built binaries
	// TODO: delete local audio output files

}

func getCLIArgs() (string, string, string, string) {
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
	return *flagInput, *flagOutput, *flagWaveForm, *flagFormat
}

func main() {
	cleanUp()

	// populate Motivic config values in memory
	initMotivicConfig()

	serve()

	// iFile, oFile, wf, oFormat := getCLIArgs()

	// convertMIDIFileToWAVFile(iFile, oFile, wf)

}
