package main

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

var port string = "8080"

const maxUploadSizeMb int64 = 10

// 10 MB expressed with bitwise operator
const maxUploadSizeBytes int64 = maxUploadSizeMb << 20
const maxRequestBodySizeBytes int64 = maxUploadSizeBytes + 512

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
	// serve html file in root
	http.Handle("/", http.StripPrefix(strings.TrimRight(path, "/"), http.FileServer(http.Dir(directory))))
	http.HandleFunc("/upload/midi", midiFileUploadHandler)
	fmt.Println("...listening at localhost:" + port)
	http.ListenAndServe(":8080", nil)
}

func saveFile(file multipart.File, handle *multipart.FileHeader, filePath string) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(filePath, data, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Upload saved locally")
}

func serveDownloadFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// tell the browser the returned content should be downloaded
	w.Header().Add("Content-Disposition", "Attachment")
	http.ServeFile(w, r, filePath)
}

// REST API to accept files for conversion
func midiFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Println(r.Method, "not accepted at upload endpoint")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	fmt.Println("MIDI File Upload Endpoint Hit")
	// 1. PARSE UPLOADED FILE
	fmt.Println("Parsing uploaded file...")
	ts := time.Now()
	// setting max memory allocation of file to 10MB the rest will be stored automatically in tmp files
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySizeBytes)
	r.ParseMultipartForm(maxUploadSizeBytes)
	midiFile, midiFileHandle, err := r.FormFile("myMIDIFile")
	if err != nil {
		fmt.Println("Error parsing the file upload")
		fmt.Println(err)
		return
	}
	defer midiFile.Close()
	fmt.Printf("Uploaded File: \t%+v at %v\n", midiFileHandle.Filename, ts)
	fmt.Printf("File Size: \t%+vkb\n", midiFileHandle.Size)
	fmt.Printf("MIME Header: \t%+v\n", midiFileHandle.Header)
	fmt.Println("Successfully uploaded file")

	// 2. SAVE UPLOADED MIDI FILE TO DISK
	inputFilePath := "tmp/midi/" + midiFileHandle.Filename
	saveFile(midiFile, midiFileHandle, inputFilePath)

	// 3. CONVERT MIDI FILE TO AUDIO FILE
	fmt.Println("Converting copied file...")
	wavFileName := r.Form.Get("wavFileName")
	waveFormName := r.Form.Get("myWaveForm")
	outputFilePath := "./output/" + wavFileName + ".wav"
	wavFilePath := convertMIDIFileToWAVFile(inputFilePath, outputFilePath, waveFormName)
	fmt.Println("Audio file written to ", wavFilePath)
	serveDownloadFile(w, r, wavFilePath)
}
