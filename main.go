package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

var (
	flagMode     = flag.String("mode", "http", "The app mode (cli or http)")
	flagInput    = flag.String("input", "", "The file to convert")
	flagFormat   = flag.String("format", "wav", "The format to convert to (wav or aiff)")
	flagOutput   = flag.String("output", "out", "The output filename")
	flagWaveForm = flag.String("waveform", "sine", "The oscillator waveform to use")
	outputDirs   = []string{"tmp/midi/", "tmp/wav/", "output"}
)

func cleanUp() {
	// clean up binaries now
	cleanUpDirs()

	// start goroutine to listen for process interruption/termination
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		// cleanup binaries on exit
		cleanUpDirs()
		os.Exit(0)
	}()
}

func cleanUpDirs() []error {
	// delete audio files from filesystem
	var errs []error
	for _, dir := range outputDirs {
		err := cleanUpDir(dir)
		if err != nil {
			fmt.Println(err)
			errs = append(errs, err)
		}
	}
	return errs
}

func cleanUpDir(dir string) error {
	names, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range names {
		os.RemoveAll(path.Join([]string{dir, entry.Name()}...))
	}
	return nil
}

func getCLIArgs() (string, string, string, string) {
	// set up CLI IO
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

func runCLIApp() {
	inputFilePath, outputFile, wf, _ := getCLIArgs()
	outputFilePath := "./output/" + outputFile + ".wav"
	convertMIDIFileToWAVFile(inputFilePath, outputFilePath, wf)
}

func main() {
	// handle binary clean up
	cleanUp()

	// populate Motivic config values in memory
	// another option is to read config values from the file at runtime
	initMotivicConfig()
	// get any CLI args
	flag.Parse()

	// process CLI input
	if *flagMode == "cli" {
		fmt.Println("...running app in CLI mode")
		runCLIApp()
	} else {
		// spin up web server
		fmt.Println("...running app in HTTP mode")
		serve()
	}
}
