package csvmapper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
)

var (
	ErrEmptyFile            = errors.New("empty csv file provided")
	ErrHeaderMismatch       = errors.New("header does not match expected columns")
	ErrUnexpectedFieldCount = errors.New("record has unexpected field count")
	ErrMapperError          = errors.New("failed to map record")
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
		if errors.Is(err, csv.ErrFieldCount) {
			return nil, fmt.Errorf("line %d: %w", lineIdx, ErrUnexpectedFieldCount)
		}
		if err != nil {
			return nil, err
		}

		convertedRecord := make([]string, 0, len(r.expectedColumns))
		for _, col := range r.expectedColumnsOrdered {
			convertedRecord = append(convertedRecord, record[r.columnNameToIndex[col]])
		}

		element, err := r.recordMapper(convertedRecord)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w:  %v", lineIdx, ErrMapperError, err)
		}

		elements = append(elements, element)
	}

	return elements, nil
}

func (r *Mapper[T]) processHeader() error {
	header, err := r.csvReader.Read()
	if err == io.EOF {
		return ErrEmptyFile
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
		return ErrHeaderMismatch
	}

	return nil
}
