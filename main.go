package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	listenAddress := flag.String("listen", "0.0.0.0:64646", "The address on which to listen [<ip>]:<port>")
	fileListFilename := flag.String("filelist", "", "The filename from which to load the file list")
	bufferSize := *flag.Int64("bufsize", 1024, "The size of the buffers to be used when reading from log files")
	flag.Parse()

	if len(*fileListFilename) == 0 {
		log.Fatal("No filelist supplied - no point starting!")
	}

	logfiles := GetFileList(fileListFilename)

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Nothing here!\n")
	})

	router.HandleFunc("/{file}", func(w http.ResponseWriter, r *http.Request) {
		if filename, ok := GetFilename(w, r, logfiles); ok {
			RetrieveFile(w, filename, r.URL.Query(), bufferSize)
		}
	})

	router.HandleFunc("/{file}/tail", func(w http.ResponseWriter, r *http.Request) {
		if filename, ok := GetFilename(w, r, logfiles); ok {
			TailFile(w, filename, r.URL.Query(), bufferSize)
		}
	})

	log.Fatal(http.ListenAndServe(*listenAddress, router))
}

func GetFileList(filename *string) map[string]string {
	logfiles := make(map[string]string)
	f, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber += 1
		parts := strings.Split(scanner.Text(), "=")
		if len(parts) != 2 {
			log.Fatalf("Error on line %d of filelist", lineNumber)
		}
		logfiles[strings.Trim(parts[0], " \t")] = strings.Trim(parts[1], " \t")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return logfiles
}

func GetFilename(w http.ResponseWriter, r *http.Request, logfiles map[string]string) (string, bool) {
	vars := mux.Vars(r)
	filename, ok := logfiles[vars["file"]]
	if !ok {
		w.WriteHeader(404)
		fmt.Fprintf(w, "File \"%s\" not found!\n", vars["file"])
		return "", false
	}
	return filename, true
}
