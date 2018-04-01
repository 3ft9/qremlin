package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var logfiles map[string]string

func main() {
	logfiles = make(map[string]string)
	logfiles["test"] = "C:/Users/stuart/test.log"

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/{file}", DownloadHandler)
	router.HandleFunc("/{file}/stream", TailHandler)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}

func GetFilename(w http.ResponseWriter, r *http.Request) (string, bool) {
	vars := mux.Vars(r)
	filename, ok := logfiles[vars["file"]]
	if !ok {
		w.WriteHeader(404)
		fmt.Fprintf(w, "File [%s] not found!", vars["file"])
		return "", false
	}
	return filename, true
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "Nothing here!")
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if filename, ok := GetFilename(w, r); ok {
		RetrieveFile(w, filename, r.URL.Query())
	}
}

func TailHandler(w http.ResponseWriter, r *http.Request) {
	if filename, ok := GetFilename(w, r); ok {
		TailFile(w, filename, r.URL.Query())
	}
}
