package storage

import "fmt"

type CsvExportable interface {
	ToSlice() []string
}

type Exportable interface {
	CsvExportable
	fmt.Stringer
}
