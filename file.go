package main

import (
	"os"
	"io"
	"net/http"
	"fmt"
	"strconv"
	"bufio"
	"strings"
	"net/url"
	"time"
	"path"
	"log"
)

func SetCommonHeaders(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", path.Base(filename)))
	w.Header().Set("Content-Type", "text/plain")
}

func RetrieveFile(w http.ResponseWriter, filename string, options url.Values) {
	SetCommonHeaders(w, filename)

	// Open the file.
	f, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprint(w, "File not found!")
		return
	}
	defer f.Close()

	// Set the default options.
	query := options.Get("query")
	numLinesStr := options.Get("n")
	numLines := int64(0)
	if len(numLinesStr) > 0 {
		if numLines, err = strconv.ParseInt(numLinesStr, 10, 64); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid number of lines: %s!", err)
			return
		}
	}

	if numLines > 0 {
		// Last n lines requested, so we need to seek back that many line feeds.
	}

	if len(query) == 0 {
		// Send without filtering.
		io.Copy(w, f)
	} else {
		var flusher http.Flusher
		var ok bool
		if flusher, ok = w.(http.Flusher); !ok {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Cannot flush, so cannot stream. Sorry.")
			return
		}

		// Check each line to see if it matches the query.
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, query) {
				fmt.Fprintf(w, "%s\n", line)
				flusher.Flush()
			}
		}
		if err := scanner.Err(); err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Failed to read file!")
			return
		}
	}
}

func TailFile(w http.ResponseWriter, filename string, options url.Values) {
	SetCommonHeaders(w, filename)

	connectionClosed := false
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		connectionClosed = true
	}()

	f, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Failed to open file!")
		return
	}
	defer f.Close()

	query := options.Get("query")
	queryLen := len(query)

	var flusher http.Flusher
	var ok bool
	if flusher, ok = w.(http.Flusher); !ok {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Cannot flush, so cannot tail. Sorry.")
		return
	}

	offset, err := f.Seek(0, io.SeekEnd)
	buffer := make([]byte, 1024, 1024)
	lineBuffer := make([]byte, 1024, 1024)
	lineBufferPos := 0
	for {
		if connectionClosed {
			return
		}

		fi, err := f.Stat()
		if err != nil {
			// Could not obtain stat, file has disappeared?
			fmt.Fprintf(w, "---------------------\nFILE HAS DISAPPEARED!\n")
			return
		}
		if fi.Size() < offset {
			offset = 0
			fmt.Fprintf(w, "(file truncated or rotated)\n")
			f.Close()
			f, err = os.Open(filename)
			if err != nil {
				fmt.Fprintf(w, "(failed to reopen file)")
				return
			}
		}

		readBytes, err := f.ReadAt(buffer, offset)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading lines:", err)
				break
			}
		}
		offset += int64(readBytes)
		if readBytes != 0 {
			thisChunk := ""
			for bufferPos := 0; bufferPos < readBytes; bufferPos++ {
				lineBuffer[lineBufferPos] = buffer[bufferPos]
				lineBufferPos += 1
				if buffer[bufferPos] == '\r' || buffer[bufferPos] == '\n' {
					if queryLen == 0 || strings.Contains(string(lineBuffer[:lineBufferPos]), query) {
						thisChunk += string(lineBuffer[:lineBufferPos])
					}
					if buffer[bufferPos] == '\r' && buffer[bufferPos+1] == '\n' {
						bufferPos += 1
					}
					lineBufferPos = 0
				}
			}
			log.Print(thisChunk)
			fmt.Fprint(w, thisChunk)
			flusher.Flush()
		}
		time.Sleep(time.Second * 2)
	}
}
