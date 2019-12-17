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

func conversionResponse(w http.ResponseWriter, outputFilePath string, fileName string) {
	tsCreated := time.Now()
	tsExpires := tsCreated.Local().Add(time.Minute * time.Duration(downloadTTLMins))
	strExpires := tsExpires.Format(time.RFC1123)
	fileURL := getAbsoluteURL("download/"+fileName, "")
	data := APIResponse{URL: fileURL, CreatedTimeStamp: tsCreated}
	var jsonData []byte
	jsonData, _ = json.MarshalIndent(data, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(strExpires)
	w.Header().Set("Expires", strExpires)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
	go expireFile(outputFilePath)
}

// REST API to accept files for conversion
// TODO: zip download file if size > threshold
// TODO: accept Motivic.json files to convert to MIDI
// TODO: accept Motivic JSON to convert to MIDI
// TODO: accept MIDI files to convert to Motivic JSON
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
	randomString := getRandomString(8)
	inputFilePath := "input/" + randomString + "_" + midiFileHandle.Filename
	saveFile(midiFile, midiFileHandle, inputFilePath)
	go expireFile(inputFilePath)

	// 3. CONVERT MIDI FILE TO AUDIO FILE
	fmt.Println("Converting copied file...")
	wavFileName := randomString + "_" + r.Form.Get("wavFileName") + ".wav"
	waveFormName := r.Form.Get("myWaveForm")
	outputFilePath := "./output/" + wavFileName
	// channel to wait for go routine response
	c := make(chan bool)
	go convertMIDIFileToWAVFile(inputFilePath, outputFilePath, waveFormName, c)
	success := <-c
	// 4. RETURN URL OF NEW FILE
	if success {
		conversionResponse(w, outputFilePath, wavFileName)
	}
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
		userFileName := strings.Split(fileName, "_")[1]
		downloadFilePath := "./output/" + fileName
		if fileExists(downloadFilePath) {
			serveDownloadFile(w, r, downloadFilePath, userFileName)
		} else {
			fmt.Println("Requested file does not exist or has expired")
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	} else {
		fmt.Println("Bad request path")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
