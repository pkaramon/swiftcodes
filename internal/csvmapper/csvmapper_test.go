package csvmapper

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPerson struct {
	Name string
	Age  int
}

func testPersonMapper(cols []string) (testPerson, error) {
	age, err := strconv.Atoi(cols[1])
	if err != nil {
		return testPerson{}, err
	}
	return testPerson{Name: cols[0], Age: age}, nil
}

func TestMapper(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		m := New(strings.NewReader(``),
			[]string{"Name", "Age"},
			func(cols []string) (testPerson, error) { return testPerson{}, nil })

		result, err := m.MapAll()
		assert.ErrorIs(t, err, ErrEmptyFile)
		assert.Nil(t, result)
	})

	t.Run("missing required header", func(t *testing.T) {
		csv := `WrongColumn,Age
John,25`
		m := New(strings.NewReader(csv),
			[]string{"Name", "Age"},
			func(cols []string) (testPerson, error) { return testPerson{}, nil })

		result, err := m.MapAll()
		assert.ErrorIs(t, err, ErrHeaderMismatch)
		assert.Nil(t, result)
	})

	t.Run("missing column in data", func(t *testing.T) {
		csv := `Name,Age
John`
		m := New(strings.NewReader(csv),
			[]string{"Name", "Age"},
			func(cols []string) (testPerson, error) { return testPerson{}, nil })

		result, err := m.MapAll()
		assert.ErrorIs(t, err, ErrUnexpectedFieldCount)
		assert.Nil(t, result)
	})

	t.Run("mapper error", func(t *testing.T) {
		csv := `Name,Age
John,25`
		m := New(strings.NewReader(csv),
			[]string{"Name", "Age"},
			func(cols []string) (testPerson, error) {
				return testPerson{}, fmt.Errorf("mapping failed")
			})

		result, err := m.MapAll()
		assert.ErrorIs(t, err, ErrMapperError)
		assert.Nil(t, result)
	})

	t.Run("successful mapping", func(t *testing.T) {
		csv := `Name,Age,Extra
John,25,ignored
Jane,30,ignored`
		m := New(strings.NewReader(csv),
			[]string{"Name", "Age"},
			testPersonMapper)

		result, err := m.MapAll()
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "John", result[0].Name)
		assert.Equal(t, 25, result[0].Age)
		assert.Equal(t, "Jane", result[1].Name)
		assert.Equal(t, 30, result[1].Age)
	})

	t.Run("columns in different order", func(t *testing.T) {
		csv := `Age,Name
25,John`
		m := New(strings.NewReader(csv),
			[]string{"Name", "Age"},
			testPersonMapper)

		result, err := m.MapAll()
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "John", result[0].Name)
		assert.Equal(t, 25, result[0].Age)
	})
}
