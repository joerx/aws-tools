package main

import (
	"encoding/csv"
	"io"
)

// RowSet is a sortable slice of rows to export. Basically a slice of ExportRows and metadata
type RowSet struct {
	Rows    [][]string
	SortCol int
	Headers []string
}

func (p RowSet) Len() int {
	return len(p.Rows)
}

func (p RowSet) Less(i, j int) bool {
	return p.Rows[i][p.SortCol] < p.Rows[j][p.SortCol]
}

func (p RowSet) Swap(i, j int) {
	p.Rows[i], p.Rows[j] = p.Rows[j], p.Rows[i]
}

// WriteToCSV writes an RowSet to a csv file at the given path
func WriteToCSV(w io.Writer, data *RowSet) error {
	cw := csv.NewWriter(w)
	cw.Write(data.Headers)
	cw.WriteAll(data.Rows)
	cw.Flush()
	return nil
}
