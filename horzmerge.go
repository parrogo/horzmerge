// Package horzmerge merges columns from one or more streams of data.
package horzmerge

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// Options struct groups all options
// accepted by Merge.
//
// Target field contains the io.Writer
// on which to write merged columns.
// When it's nil, os.Stdout is used as writer.
//
// Empty field is a string that can be used
// to specify a value that will be interpreted as
// empty. When a cell of text contains this
// value (space trimmed), it cannot overwrite a
// previous value already set for the column,
// readed from one of the previous readers.
type Options struct {
	Target io.Writer
	Empty  string
}

// Merge read lines from all io.Reader in sources,
// and build an hash of columns for every reader,
// interpreting data with a tabular semantic.
//
// All hashes created in this way are then merged
// into a single hash and saved in a tabular format
// again.
func Merge(opt Options, readers ...io.Reader) error {
	if len(readers) == 0 {
		return errors.New("no source readers provided")
	}

	var out *bufio.Writer
	if opt.Target != nil {
		out = bufio.NewWriter(opt.Target)
	} else {
		out = bufio.NewWriter(os.Stdout)
	}

	sources := make([]map[string]string, len(readers))

	headerOrder := map[string]int{}

	for idx, r := range readers {
		source := bufio.NewReader(r)
		headers, err := readHeaders(source)
		for _, h := range headers {
			if _, exists := headerOrder[h]; !exists {
				headerOrder[h] = len(headerOrder)
			}
		}
		if err != nil {
			return fmt.Errorf("error reading from source %d: %w", idx, InputError{err, idx})
		}
		values, err := readValues(source)
		if err != nil {
			return fmt.Errorf("error reading from source %d: %w", idx, InputError{err, idx})
		}

		hash := map[string]string{}
		for idx, head := range headers {
			val := values[idx]
			hash[head] = val
		}
		sources[idx] = hash
	}

	merged := map[string]string{}
	for _, hash := range sources {
		for key, val := range hash {
			if mv, exists := merged[key]; !exists || strings.TrimSpace(mv) == opt.Empty {
				merged[key] = val
			}
		}
	}

	var werr error

	write := func(s string) {
		if werr != nil {
			return
		}
		_, e := out.WriteString(s)
		if e != nil {
			werr = fmt.Errorf("error writing output: %w", e)
		}
	}

	headers := make([]string, len(merged))
	values := make([]string, len(merged))

	for k, v := range merged {
		idx := headerOrder[k]
		headers[idx] = k
		values[idx] = v
		idx++
	}

	for _, h := range headers {
		write(h)
	}

	write("\n")

	for _, v := range values {
		write(v)
	}

	write("\n")

	if werr == nil {
		e := out.Flush()
		if e != nil {
			werr = fmt.Errorf("error flushing output: %w", e)
		}
	}
	return werr
}

func checkHeaders(source *bufio.Reader, headers []string) error {
	sourceHeaders, err := readHeaders(source)
	if err != nil {
		return err
	}
	if len(sourceHeaders) != len(headers) {
		return fmt.Errorf("headers len differs: expected %d, got %d", len(headers), len(sourceHeaders))
	}
	for idx, head := range headers {
		sourceHead := sourceHeaders[idx]
		if head != sourceHead {
			return fmt.Errorf("field header %d differs: expected `%s`, got `%s`", idx, head, sourceHead)
		}
	}
	return nil
}

func readHeaders(source *bufio.Reader) ([]string, error) {
	return readValues(source)
}

func readValues(source *bufio.Reader) ([]string, error) {
	reader := source
	headers := []string{}
	var val string
	var len int

	emit := func() {
		h := fmt.Sprintf("%*s", len, val)
		headers = append(headers, h)
		len = 1
		val = ""
	}

	var r rune
	var err error
	for {
		r, _, err = reader.ReadRune()
		if err == io.EOF || r == '\n' {
			if val != "" {
				emit()
			}
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading headers: %w", err)
		}

		if unicode.IsSpace(r) {
			if val == "" {
				len++
			} else {
				emit()
			}

			continue

		} else {
			len++
		}

		val += string(r)
	}

	return headers, nil
}

// InputError wraps an error
// in order to include the position
// of failing stream.
type InputError struct {
	err error
	idx int
}

// Error implements error interface
func (e InputError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error
func (e InputError) Unwrap() error {
	return e.err
}

// Convert returns an error that include the
// name of the file that causes the failure
func (e InputError) Convert(filenames []string) error {
	return fmt.Errorf("Cannot read file %s: %w", filenames[e.idx], e.err)
}
