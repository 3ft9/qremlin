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
)

func SetCommonHeaders(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "text/plain")
}

func RetrieveFile(w http.ResponseWriter, filename string, options url.Values, bufferSize int64) {
	SetCommonHeaders(w, filename)

	// Open the file.
	f, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprint(w, "File not found!\n")
		return
	}
	defer f.Close()

	// Set the default options.
	query := options.Get("q")
	numLinesStr := options.Get("n")
	numLines := int64(0)
	if len(numLinesStr) > 0 {
		if numLines, err = strconv.ParseInt(numLinesStr, 10, 64); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid number of lines: %s!\n", err)
			return
		}
	}

	if numLines > 0 {
		buffer := make([]byte, bufferSize)
		offset, err := f.Seek(-bufferSize, 2)
		if err != nil {
			offset, err = f.Seek(0, 0)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to find the end of the file: %s\n", err)
				return
			}
		}

		for numLines > 0 {
			if offset == 0 {
				offset, err = f.Seek(0, 1)
				break
			}
			readBytes, err := f.ReadAt(buffer, offset)
			if readBytes == 0 {
				break
			}
			if err != nil && err != io.EOF {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to read part of the file: %s\n", err)
				return
			}
			for i := readBytes - 1; i >= 0; i-- {
				if buffer[i] == '\n' {
					numLines -= 1
					if numLines < 0 {
						offset, err = f.Seek(int64(i)+1, 1)
						break
					}
				}
			}

			if numLines > 0 {
				offset, err = f.Seek(-bufferSize, 1)
			}
		}
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", path.Base(filename)))

	if len(query) == 0 {
		// Send without filtering.
		io.Copy(w, f)
	} else {
		var flusher http.Flusher
		var ok bool
		if flusher, ok = w.(http.Flusher); !ok {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Cannot flush, so cannot stream. Sorry.\n")
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
			fmt.Fprint(w, "Failed to read file!\n")
			return
		}
	}
}

func TailFile(w http.ResponseWriter, filename string, options url.Values, bufferSize int64) {
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
		fmt.Fprintf(w, "Failed to open file!\n")
		return
	}
	defer f.Close()

	query := options.Get("q")
	queryLen := len(query)

	var flusher http.Flusher
	var ok bool
	if flusher, ok = w.(http.Flusher); !ok {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Cannot flush, so cannot tail. Sorry.\n")
		return
	}

	offset, err := f.Seek(0, 2)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Failed to seek to the end of the file.\n")
		return
	}
	buffer := make([]byte, bufferSize, bufferSize)
	lineBuffer := make([]byte, bufferSize, bufferSize)
	lineBufferPos := 0
	for {
		if connectionClosed {
			return
		}

		fi, err := f.Stat()
		if err != nil {
			fmt.Fprintf(w, "---------------------\nFILE HAS DISAPPEARED!\n")
			return
		}
		if fi.Size() < offset {
			offset = 0
			fmt.Fprintf(w, "(file truncated or rotated)\n")
			f.Close()
			f, err = os.Open(filename)
			if err != nil {
				fmt.Fprintf(w, "(failed to reopen file)\n")
				return
			}
		}

		readBytes, err := f.ReadAt(buffer, offset)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(w, "(error reading from the file)\n")
				return
			}
		}
		offset += int64(readBytes)
		if readBytes != 0 {
			thisChunk := ""
			for bufferPos := 0; bufferPos < readBytes; bufferPos++ {
				lineBuffer[lineBufferPos] = buffer[bufferPos]
				lineBufferPos += 1
				if buffer[bufferPos] == '\n' {
					if queryLen == 0 || strings.Contains(string(lineBuffer[:lineBufferPos]), query) {
						thisChunk += string(lineBuffer[:lineBufferPos])
					}
					lineBufferPos = 0
				}
			}
			if len(thisChunk) > 0 {
				fmt.Fprint(w, thisChunk)
				flusher.Flush()
			}
		}
		time.Sleep(time.Second)
	}
}
