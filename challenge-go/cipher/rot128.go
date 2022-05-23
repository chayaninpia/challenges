package cipher

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
)

const SLASH_R = 13 // \r
const SLASH_N = 10 // \n
// Rot128Reader implements io.Reader that transforms
type Rot128Reader struct {
	reader  io.Reader
	scanner *bufio.Scanner
	file    *os.File
}

func NewRot128Reader(path string) (*Rot128Reader, error) {
	csvOpen, err := os.Open(path)
	if err != nil {
		csvOpen.Close()
		log.Fatal(err.Error())
	}

	reader := bufio.NewReader(csvOpen)
	if reader == nil {
		csvOpen.Close()
		log.Fatal("io.Reader is nil")
	}
	s := bufio.NewScanner(reader)
	s.Split(scanLinesRot128)
	return &Rot128Reader{
		reader:  csvOpen,
		scanner: s,
		file:    csvOpen}, nil
}

func (r *Rot128Reader) Scan() (string, bool) {
	if ok := r.scanner.Scan(); !ok {
		return ``, false
	}
	data := r.scanner.Bytes()
	rot128(data)
	return string(data), true
}

func (r *Rot128Reader) Close() {
	r.file.Close()
}

func (r *Rot128Reader) Read(p []byte) (int, error) {
	if n, err := r.reader.Read(p); err != nil {
		return n, err
	} else {
		rot128(p[:n])
		return n, nil
	}
}

type Rot128Writer struct {
	writer io.Writer
	buffer []byte // not thread-safe
}

func NewRot128Writer(w io.Writer) (*Rot128Writer, error) {
	return &Rot128Writer{
		writer: w,
		buffer: make([]byte, 4096, 4096),
	}, nil
}

func (w *Rot128Writer) Write(p []byte) (int, error) {
	n := copy(w.buffer, p)
	rot128(w.buffer[:n])
	return w.writer.Write(w.buffer[:n])
}

func rot128(buf []byte) {
	for idx := range buf {
		buf[idx] += 128
	}
}

// standard function from  https://golang.org/src/bufio/scan.go?s=11799:11877#L335
// and adapt to check the rot128
// CSV files are not necessarily ending with \n, Microsoft format will be \r\n
// that's why they implement their dropCR function

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == SLASH_R+128 {
		return data[0 : len(data)-1]
	}
	return data
}

func scanLinesRot128(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, SLASH_N+128); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}
