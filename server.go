package main

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

var port string = "8080"

// FileSystem custom file system handler
type FileSystem struct {
	fs http.FileSystem
}

// Open opens file
func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}
func serve() {
	directory := "./"
	path := "/"
	http.Handle("/", http.StripPrefix(strings.TrimRight(path, "/"), http.FileServer(http.Dir(directory))))
	// http.HandleFunc("/", staticFileServer)
	http.HandleFunc("/upload/midi", midiFileUploadHandler)
	fmt.Println("...listening at localhost:" + port)
	http.ListenAndServe(":8080", nil)
}

func staticFileServer(w http.ResponseWriter, r *http.Request) {
	directory := "./"
	fileServer := http.FileServer(FileSystem{http.Dir(directory)})
	http.StripPrefix(strings.TrimRight(directory, "/"), fileServer)
}

func writeTempFile(dir string, name string) (*os.File, error) {
	// copies content of a file to a temp file
	tempFile, err := ioutil.TempFile(dir, name)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer tempFile.Close()
	return tempFile, nil
}

func copyFile(source *os.File, target *os.File) (*os.File, error) {
	// read the contents of source file into byte array
	fileBytes, err := ioutil.ReadAll(source)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// write this byte array to target file
	target.Write(fileBytes)
	target.Close()
	source.Close()
	return target, nil
}
func copyMultiPartFile(source multipart.File, target *os.File) (*os.File, error) {
	// read the contents of source file into byte array
	fileBytes, err := ioutil.ReadAll(source)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// write this byte array to target file
	target.Write(fileBytes)
	target.Close()
	source.Close()
	return target, nil
}

func serveDownloadFile(w http.ResponseWriter, r *http.Request, ts time.Time, filePath string) {
	http.ServeFile(w, r, filePath)

	// // ServeContent uses the name for mime detection
	// const name = "random.txt"
	// // tell the browser the returned content should be downloaded
	// w.Header().Add("Content-Disposition", "Attachment")
	// http.ServeContent(w, r, name, ts, content)
}

// REST API to accept files for conversion
func midiFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("MIDI File Upload Endpoint Hit")
	ts := time.Now()
	r.ParseMultipartForm(10 << 20)
	midiFile, handler, err := r.FormFile("myMIDIFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer midiFile.Close()
	fmt.Printf("Uploaded File: %+v at %v\n", handler.Filename, ts)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempMIDIFile, err := writeTempFile("tmp/midi", fmt.Sprintf("upload-*.midi"))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	tempMIDIFile, err = copyMultiPartFile(midiFile, tempMIDIFile)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
	fmt.Println("tempFilePath:", tempMIDIFile.Name())
	wavFileName := r.Form.Get("wavFileName")
	waveFormName := r.Form.Get("myWaveForm")
	wavFilePath := convertMIDIFileToWAVFile(tempMIDIFile.Name(), wavFileName, waveFormName)
	tempWAVFile, err := writeTempFile("tmp/wav", fmt.Sprintf("%s.wav", wavFileName))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	newWavFile, err := os.Open(wavFilePath)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	tempWAVFile, err = copyFile(newWavFile, tempWAVFile)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	serveDownloadFile(w, r, ts, tempWAVFile.Name())

}
