package jsonz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

// LineParseStream reads line-delimited JSON from r and streams decoded elements to fn.
// The callback fn returns a boolean (true to keep streaming, false to stop) and an error.
// If the callback returns an error, the streaming halts and the error is returned.
func LineParseStream[E any](r io.Reader, fn func(item E) (bool, error)) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	for {
		var item E
		err := decoder.Decode(&item)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to parse line: %w", err)
		}
		keepGoing, cbErr := fn(item)
		if cbErr != nil {
			return cbErr
		}
		if !keepGoing {
			break
		}
	}
	return nil
}

// LineParseReader reads line-delimited JSON from r and appends the decoded elements to v.
func LineParseReader[S ~[]E, E any](r io.Reader, v *S) error {
	if v == nil {
		return fmt.Errorf("destination slice pointer cannot be nil")
	}
	return LineParseStream(r, func(item E) (bool, error) {
		*v = append(*v, item)
		return true, nil
	})
}

// LineParseInto unmarshals line-delimited JSON data from blobz.Data into v.
func LineParseInto[S ~[]E, E any, I blobz.Data](data I, v *S) error {
	err := LineParseReader(blobz.New(data).NewReader(), v)
	if err != nil {
		return fmt.Errorf("failed to parse into slice: %w", err)
	}
	return nil
}

// LineFormatWriter encodes each element in v as a JSON line and writes it to w.
func LineFormatWriter[S ~[]E, E any](w io.Writer, v S) error {
	encoder := json.NewEncoder(w)
	for i, item := range v {
		err := encoder.Encode(item)
		if err != nil {
			return fmt.Errorf("failed to encode item at index %d: %w", i, err)
		}
	}
	return nil
}

// LineFormatReadCloser returns an io.ReadCloser that provides the line-delimited JSON-encoded
// representation of v. It streams the output and does not load the entire JSON object into memory.
// The caller must close the reader when finished.
func LineFormatReadCloser[S ~[]E, E any](v S) io.ReadCloser {
	r, w := io.Pipe()

	go func() {
		var err error
		defer func() {
			w.CloseWithError(err)
		}()
		err = LineFormatWriter(w, v)
	}()

	return r
}

// LineFormatReader streams the line-delimited JSON-encoded representation of v to the given callback fn.
// It automatically closes the underlying stream once fn completes.
func LineFormatReader[S ~[]E, E any](v S, fn func(r io.Reader) error) error {
	r := LineFormatReadCloser(v)
	defer r.Close()
	err := fn(r)
	if err != nil {
		return fmt.Errorf("callback failed: %w", err)
	}
	return nil
}

// LineFormat marshals the elements of v into line-delimited JSON and returns it as a blobz.Blob.
func LineFormat[S ~[]E, E any](v S) (blobz.Blob, error) {
	var buf bytes.Buffer
	err := LineFormatWriter(&buf, v)
	if err != nil {
		return nil, fmt.Errorf("failed to format slice: %w", err)
	}
	return blobz.New(buf.Bytes()), nil
}

// LineFormatBytes marshals the elements of v into line-delimited JSON and returns it as a byte slice.
func LineFormatBytes[S ~[]E, E any](v S) ([]byte, error) {
	ret, err := LineFormat(v)
	if err != nil {
		return nil, fmt.Errorf("failed to format bytes: %w", err)
	}
	return ret.Bytes(), nil
}
