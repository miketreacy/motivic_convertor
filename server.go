package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
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
const inputFileDir string = "./input/"
const outputFileDir string = "./output/"
const maxUploadSizeMb int64 = 10

// 10 MB expressed with bitwise operator
const maxUploadSizeBytes int64 = maxUploadSizeMb << 20
const maxRequestBodySizeBytes int64 = maxUploadSizeBytes + 512

const downloadTTLMins int64 = 1

// APIResponse : response for /download/<filename>
type APIResponse struct {
	URL              string    `json:"url"`
	CreatedTimeStamp time.Time `json:"created"`
	Message          string    `json:"message"`
	Success          bool      `json:"success"`
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

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func zipFiles(filename string, files []string, key string) error {
	zipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file, key); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath string, key string) error {
	fileToZip, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file.
	// To reflect dir structure, set this to the full path.
	header.Name = getFileNameFromPath(filePath, key)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
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
	// ignore error if dir already exists
	_ = os.Mkdir(inputFileDir, 0777)
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
	data := APIResponse{}
	tsCreated := time.Now()
	// conversion failed
	if outputFilePath == "" {
		data = APIResponse{URL: "", CreatedTimeStamp: tsCreated, Success: false, Message: "Conversion failed"}
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		tsExpires := tsCreated.Local().Add(time.Minute * time.Duration(downloadTTLMins))
		strExpires := tsExpires.Format(time.RFC1123)
		fileURL := getAbsoluteURL("download/"+fileName, "")
		data = APIResponse{URL: fileURL, CreatedTimeStamp: tsCreated, Success: true, Message: "File converted"}
		fmt.Println(strExpires)
		w.Header().Set("Expires", strExpires)
		w.WriteHeader(http.StatusOK)
	}
	var jsonData []byte
	jsonData, _ = json.MarshalIndent(data, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
	go expireFile(outputFilePath)
}

// REST API to accept files for conversion
// TODO: handle polyphonic MIDI - support or return helpful exception response
// TODO: increase conversion types:
// 		Motivic.json file => MIDI
// 		Motivic JSON payload => MIDI
// 		Motivic JSON payload => WAV
// 		MIDI files => Motivic JSON response
// 		MIDI files => Motivic.json file
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
	inputFilePath := inputFileDir + randomString + "_" + midiFileHandle.Filename
	saveFile(midiFile, midiFileHandle, inputFilePath)
	go expireFile(inputFilePath)

	// 3. CONVERT MIDI FILE TO AUDIO FILE
	fmt.Println("Converting copied file...")
	waveFormName := r.Form.Get("myWaveForm")
	outputFileName := r.Form.Get("wavFileName")
	wavFileoutputFilePath, _ := getFilePathFromName(outputFileDir, randomString, outputFileName, "wav")
	// channel to wait for go routine response
	c := make(chan bool)
	go convertMIDIFileToWAVFile(inputFilePath, wavFileoutputFilePath, waveFormName, c)
	success := <-c
	go expireFile(wavFileoutputFilePath)

	// 4. RETURN URL OF NEW FILE
	var zipFileOutputPath string = ""
	var zipFileName string = ""
	if success {
		zipFileOutputPath, zipFileName = getFilePathFromName(outputFileDir, randomString, outputFileName, "zip")
		filesToZip := []string{wavFileoutputFilePath}
		if err := zipFiles(zipFileOutputPath, filesToZip, randomString); err != nil {
			panic(err)
		}
		fmt.Println("Zipped File:", zipFileOutputPath)

	}
	conversionResponse(w, zipFileOutputPath, zipFileName)
}

func fileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	fileName := paths[2:len(paths)]
	if len(fileName) == 1 {
		fileName := fileName[0]
		userFileName := strings.Split(fileName, "_")[1]
		downloadFilePath := outputFileDir + fileName
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
