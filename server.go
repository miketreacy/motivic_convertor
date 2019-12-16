package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const protocol string = "http"
const domain string = "localhost"
const port string = "8080"

const maxUploadSizeMb int64 = 10

// 10 MB expressed with bitwise operator
const maxUploadSizeBytes int64 = maxUploadSizeMb << 20
const maxRequestBodySizeBytes int64 = maxUploadSizeBytes + 512

const downloadTTLMins int64 = 1

// APIResponse : response for /download/<filename>
type APIResponse struct {
	URL              string    `json:"url"`
	CreatedTimeStamp time.Time `json:"created"`
}

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
	http.HandleFunc("/download/", fileDownloadHandler)
	fmt.Println("...listening at " + domain + ":" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getAbsoluteURL(path string, query string) string {
	qs := ""
	if query != "" {
		qs = "?" + query
	}
	return fmt.Sprintf("%v://%v:%v/%v%v", protocol, domain, port, path, qs)
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

func serveDownloadFile(w http.ResponseWriter, r *http.Request, filePath string, fileName string) {
	// tell the browser the returned content should be downloaded
	contentDisposition := fmt.Sprintf("attachment; filename=\"%v\"", fileName)
	w.Header().Add("Content-Disposition", contentDisposition)
	http.ServeFile(w, r, filePath)
}

// REST API to accept files for conversion
// TODO: convert this to async go routine
// TODO: zip download file if size > threshold
// TODO: delete files after downloadTTLMins
func midiFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Println(r.Method, "not accepted at upload endpoint")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	fmt.Println("MIDI File Upload Endpoint Hit")
	// 1. PARSE UPLOADED FILE
	fmt.Println("Parsing uploaded file...")
	tsCreated := time.Now()
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
	fmt.Printf("Uploaded File: \t%+v at %v\n", midiFileHandle.Filename, tsCreated)
	fmt.Printf("File Size: \t%+vkb\n", midiFileHandle.Size)
	fmt.Printf("MIME Header: \t%+v\n", midiFileHandle.Header)
	fmt.Println("Successfully uploaded file")

	// 2. SAVE UPLOADED MIDI FILE TO DISK
	inputFilePath := "tmp/midi/" + midiFileHandle.Filename
	saveFile(midiFile, midiFileHandle, inputFilePath)

	// 3. CONVERT MIDI FILE TO AUDIO FILE
	fmt.Println("Converting copied file...")
	wavFileName := r.Form.Get("wavFileName") + ".wav"
	waveFormName := r.Form.Get("myWaveForm")
	outputFilePath := "./output/" + wavFileName
	convertMIDIFileToWAVFile(inputFilePath, outputFilePath, waveFormName)
	fmt.Println("WAV File created at", outputFilePath)

	// 4. RETURN URL OF NEW FILE
	tsCreated = time.Now()
	tsExpires := tsCreated.Local().Add(time.Minute * time.Duration(downloadTTLMins))
	strExpires := tsExpires.Format(time.RFC1123)
	fileURL := getAbsoluteURL("download/"+wavFileName, "")
	data := APIResponse{URL: fileURL, CreatedTimeStamp: tsCreated}
	var jsonData []byte
	jsonData, err = json.MarshalIndent(data, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(strExpires)
	w.Header().Set("Expires", strExpires)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
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

func fileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	fileName := paths[2:len(paths)]
	if len(fileName) == 1 {
		fileName := fileName[0]
		downloadFilePath := "./output/" + fileName
		if fileExists(downloadFilePath) {
			serveDownloadFile(w, r, downloadFilePath, fileName)
		} else {
			fmt.Println("Requested file does not exist (or is a directory)")
			return
		}

	} else {
		fmt.Println("Bad request path")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}
