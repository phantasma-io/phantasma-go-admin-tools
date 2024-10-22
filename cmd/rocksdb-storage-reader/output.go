package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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
	format     OutputFormat
	csv        *csv.Writer
	records    []string
	AnyRecords []any
}

func (o *Output) Init(format OutputFormat) {
	o.format = format

	if o.format == CSV {
		o.csv = csv.NewWriter(io.Writer(os.Stdout))
		o.records = []string{}
		o.AnyRecords = []any{}
	} else if o.format == JSON {
		o.records = []string{}
		o.AnyRecords = []any{}
	}
}

func NewOutput(format OutputFormat) *Output {
	var o = new(Output)

	o.Init(format)
	return o
}

func (o *Output) AddStringRecord(r string) {
	if o.format == CSV || o.format == JSON {
		o.records = append(o.records, r)
	} else if o.format == PLAIN {
		fmt.Println(r)
	}
}

func (o *Output) AddAnyRecord(r fmt.Stringer) {
	if o.format == CSV {
		o.records = append(o.records, r.String())
	} else if o.format == JSON {
		o.AnyRecords = append(o.AnyRecords, r)
	} else if o.format == PLAIN {
		fmt.Println(r)
	}
}

func (o *Output) Flush() {
	if o.format == CSV {
		o.csv.Write([]string(o.records))

		o.csv.Flush()
	} else if o.format == JSON {
		var row []byte
		var err error
		if len(o.records) > 0 {
			row, err = json.Marshal(o.records)
		} else {
			row, err = json.Marshal(o.AnyRecords)
		}

		if err != nil {
			panic(err)
		}
		fmt.Println(string(row))
	}
}
