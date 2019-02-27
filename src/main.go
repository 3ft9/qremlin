package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

var VERSION string = "3"

func main() {
	showVersion := flag.Bool("version", false, "Show the version number and exit")
	listenAddress := flag.String("listen", "0.0.0.0:64646", "The address on which to listen [<ip>]:<port>")
	fileListFilename := flag.String("filelist", "/etc/qremlin-filelist.conf", "The filename from which to load the file list")
	bufferSize := *flag.Int64("bufsize", 10485760, "The size of the buffers to be used when reading from log files")
	flag.Parse()

	if *showVersion {
		fmt.Printf("qremlin v%s\n", VERSION)
		os.Exit(1)
	}

	fmt.Printf("Config:\n")
	fmt.Printf("    listen = %s\n", *listenAddress)
	fmt.Printf("  filelist = %s\n", *fileListFilename)
	fmt.Printf("   bufsize = %d\n", bufferSize)
	fmt.Printf("\n")

	logfiles := GetFileList(fileListFilename)

	OutputLogfiles(os.Stdout, logfiles)
	fmt.Printf("\n")

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		OutputLogfiles(w, logfiles)
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

	fmt.Printf("Server started on %s\n", *listenAddress)

	log.Fatal(http.ListenAndServe(*listenAddress, router))
}

func OutputLogfiles(w io.Writer, logfiles map[string]string) {
	fmt.Fprintf(w, "Available log files:\n")
    keys := make([]string, len(logfiles))
	maxidlen := 0
	i := 0
	for k, _ := range logfiles {
		if len(k) > maxidlen {
			maxidlen = len(k)
		}
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		fmt.Fprintf(w, "%s%s : %s\n", strings.Repeat(" ", maxidlen - len(keys[i])), keys[i], logfiles[keys[i]])
	}
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
