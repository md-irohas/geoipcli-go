package main

import (
	"encoding/csv"
	"os"
	"strings"
)

type Writer struct {
	w                 *csv.Writer
	EscapeComma       bool
	EscapeDoubleQuote bool
}

func NewCSVWriter() *Writer {
	w := &Writer{w: csv.NewWriter(os.Stdout), EscapeComma: true, EscapeDoubleQuote: true}
	w.w.Comma = ','
	return w
}

func NewTSVWriter() *Writer {
	w := &Writer{w: csv.NewWriter(os.Stdout), EscapeComma: true, EscapeDoubleQuote: true}
	w.w.Comma = '\t'
	return w
}

func (w *Writer) Write(record []string) {
	if w.EscapeComma {
		for i, s := range record {
			record[i] = strings.Replace(s, ",", "<comma>", -1)
		}
	}
	if w.EscapeDoubleQuote {
		for i, s := range record {
			record[i] = strings.Replace(s, "\"", "<doublequotes>", -1)
		}
	}

	w.w.Write(record)
	w.w.Flush()
}
