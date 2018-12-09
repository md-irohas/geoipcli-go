package main

import (
	"encoding/csv"
	"io"
	"strings"
)

type Writer struct {
	w                  *csv.Writer
	EscapeComma        bool
	EscapeDoubleQuotes bool
}

func NewCSVWriter(w io.Writer) *Writer {
	return &Writer{w: csv.NewWriter(w), EscapeComma: true, EscapeDoubleQuotes: true}
}

func (w *Writer) Write(record []string) {
	if w.EscapeComma || w.EscapeDoubleQuotes {
		if w.EscapeComma {
			for i, s := range record {
				record[i] = strings.Replace(s, ",", "<comma>", -1)
			}
		}
		if w.EscapeDoubleQuotes {
			for i, s := range record {
				record[i] = strings.Replace(s, "\"", "<doublequotes>", -1)
			}
		}
	}
	w.w.Write(record)
	w.w.Flush()
}
