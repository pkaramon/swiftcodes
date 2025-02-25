package csvmapper

import (
	"encoding/csv"
	"fmt"
	"io"
)

type Mapper[T any] struct {
	csvReader              *csv.Reader
	expectedColumns        map[string]bool
	expectedColumnsOrdered []string
	columnNameToIndex      map[string]int
	recordMapper           func([]string) (T, error)
}

func New[T any](reader io.Reader, columns []string, recordMapper func([]string) (T, error)) *Mapper[T] {
	expectedColumns := map[string]bool{}
	for _, col := range columns {
		expectedColumns[col] = true
	}

	return &Mapper[T]{
		csvReader:              csv.NewReader(reader),
		expectedColumns:        expectedColumns,
		expectedColumnsOrdered: columns,
		columnNameToIndex:      map[string]int{},
		recordMapper:           recordMapper,
	}
}

func (r *Mapper[T]) MapAll() ([]T, error) {
	if err := r.processHeader(); err != nil {
		return nil, err
	}

	elements := make([]T, 0)
	for lineIdx := 1; ; lineIdx++ {
		record, err := r.csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record %d: %v", lineIdx, err)
		}

		convertedRecord := make([]string, 0, len(r.expectedColumns))
		for _, col := range r.expectedColumnsOrdered {
			if len(record) <= r.columnNameToIndex[col] {
				return nil, fmt.Errorf("record %d does not have column %s", lineIdx, col)
			}

			convertedRecord = append(convertedRecord, record[r.columnNameToIndex[col]])
		}

		element, err := r.recordMapper(convertedRecord)
		if err != nil {
			return nil, fmt.Errorf("failed to convert record %d: %v", lineIdx, err)
		}

		elements = append(elements, element)
	}

	return elements, nil
}

func (r *Mapper[T]) processHeader() error {
	header, err := r.csvReader.Read()
	if err == io.EOF {
		return fmt.Errorf("empty csv file provided")
	}
	if err != nil {
		return err
	}

	for idx, col := range header {
		if _, ok := r.expectedColumns[col]; ok {
			r.columnNameToIndex[col] = idx
		}
	}

	if len(r.columnNameToIndex) != len(r.expectedColumns) {
		return fmt.Errorf("header does not match expected columns")
	}

	return nil
}
