package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/go-audio/midi"
)

var (
	flagMode     = flag.String("mode", "http", "The app mode (cli or http)")
	flagInput    = flag.String("input", "", "The file to convert")
	flagFormat   = flag.String("format", "wav", "The format to convert to (wav or aiff)")
	flagOutput   = flag.String("output", "out", "The output filename")
	flagWaveForm = flag.String("waveform", "sine", "The oscillator waveform to use")
	outputDirs   = []string{"input", "output"}
)

func printReflectionInfo(t *midi.Track) {
	// expect CustomStruct if non pointer
	fmt.Println("Actual type is:", reflect.TypeOf(t))

	// expect struct if non pointer
	fmt.Println("Value type is:", reflect.ValueOf(t).Kind())

	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		// expect: CustomStruct
		fmt.Println("Indirect type is:", reflect.Indirect(reflect.ValueOf(t)).Kind()) // prints interface

		// expect: struct
		fmt.Println("Indirect value type is:", reflect.Indirect(reflect.ValueOf(t)).Kind()) // prints interface
	}

	fmt.Println("")
}

func getRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func getFilePathFromName(dir string, key string, fileName string, fileType string) (string, string) {
	encodedName := key + "_" + fileName + "." + fileType
	encodedFilePath := dir + encodedName
	return encodedFilePath, encodedName
}

func getFileNameFromPath(path string, key string) string {
	return strings.Split(path, key+"_")[1]
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func expireFile(filePath string) {
	if filePath == "" || !fileExists(filePath) {
		return
	}
	fileTimer := time.NewTimer(60 * time.Second)
	<-fileTimer.C
	fmt.Println("File deleted:", filePath)
	os.Remove(filePath)
}

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
	c := make(chan bool)
	go convertMIDIFileToWAVFile(inputFilePath, outputFilePath, wf, c)
	<-c
	go expireFile(inputFilePath)
	go expireFile(outputFilePath)
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
