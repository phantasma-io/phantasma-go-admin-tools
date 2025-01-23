package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/phantasma/storage"
)

type OutputFormat int

const (
	CSV   OutputFormat = iota
	JSON  OutputFormat = iota
	PLAIN OutputFormat = iota
)

var outputFormatLookup = map[OutputFormat]string{
	CSV:   `CSV`,
	JSON:  `JSON`,
	PLAIN: `PLAIN`,
}

func OutputFormatFromString(outputFormat string) OutputFormat {
	if outputFormat == "" {
		return PLAIN
	}

	for f, s := range outputFormatLookup {
		if strings.EqualFold(s, outputFormat) {
			return f
		}
	}

	panic("Output format " + outputFormat + " not recognized")
}

type Output struct {
	format      OutputFormat
	csv         *csv.Writer
	jsonRecords []string
	csvRecords  [][]string
	AnyRecords  []any
	outputFile  *os.File
}

func (o *Output) Init(format OutputFormat) {
	o.format = format

	if appOpts.Output != "" {
		var err error
		o.outputFile, err = os.OpenFile(appOpts.Output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
	}

	if o.format == CSV {
		if appOpts.Output != "" {
			o.csv = csv.NewWriter(io.Writer(o.outputFile))
		} else {
			o.csv = csv.NewWriter(io.Writer(os.Stdout))
		}
		o.csvRecords = [][]string{}
		o.AnyRecords = []any{}
	} else if o.format == JSON {
		o.jsonRecords = []string{}
		o.AnyRecords = []any{}
	}
}

// TODO call it
func (o *Output) Uninit() {
	if appOpts.Output != "" && o.outputFile != nil {
		o.outputFile.Close()
	}
}

func NewOutput(format OutputFormat) *Output {
	var o = new(Output)

	o.Init(format)
	return o
}

func (o *Output) AddRecord(r storage.Exportable) {
	if o.format == CSV {
		o.csvRecords = append(o.csvRecords, r.ToSlice())
	} else if o.format == JSON {
		o.AnyRecords = append(o.AnyRecords, r)
	} else if o.format == PLAIN {
		fmt.Println(r)
	}
}

func (o *Output) Flush() {
	if o.format == CSV {
		o.csv.WriteAll(o.csvRecords)

		o.csv.Flush()
	} else if o.format == JSON {
		var row []byte
		var err error
		if len(o.jsonRecords) > 0 {
			row, err = json.Marshal(o.jsonRecords)
		} else {
			row, err = json.Marshal(o.AnyRecords)
		}

		if err != nil {
			panic(err)
		}

		if appOpts.Output != "" {
			o.outputFile.WriteString(string(row))
		} else {
			fmt.Println(string(row))
		}
	}
}
