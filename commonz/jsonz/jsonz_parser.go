package jsonz

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

func ParseReader[T any](r io.Reader, v T) (T, error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	err := decoder.Decode(v)
	if err != nil {
		return v, fmt.Errorf("failed to parse json: %w", err)
	}
	return v, nil
}

func Parse[T any, I blobz.Data](data I, v T) (T, error) {
	return ParseReader(blobz.New(data).NewReader(), v)
}

// ParseReaderUntilEOF reads all JSON objects from a reader until it encounters an
// io.EOF. This is useful for reading streams of JSON objects.
func ParseReaderUntilEOF[T any](r io.Reader) ([]*T, error) {
	result := make([]*T, 0)
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	for {
		var v T
		err := decoder.Decode(&v)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse json stream: %w", err)
		}
		result = append(result, &v)
	}

	return result, nil
}

func FormatWriter(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(v)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	return nil
}

// FormatReadCloser returns an io.ReadCloser that provides the JSON-encoded
// representation of v. It streams the output and does not load the entire
// JSON object into memory. The caller must close the reader when finished.
func FormatReadCloser(v any) io.ReadCloser {
	r, w := io.Pipe()

	go func() {
		var err error
		defer func() {
			// Ensure the writer is closed with the encoding error, if any.
			w.CloseWithError(err)
		}()
		encoder := json.NewEncoder(w)
		encodeErr := encoder.Encode(v)
		if encodeErr != nil {
			err = fmt.Errorf("failed to marshal data: %w", encodeErr)
		}
	}()

	return r
}

func Format(v any) (blobz.Blob, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return blobz.New(b), nil
}
