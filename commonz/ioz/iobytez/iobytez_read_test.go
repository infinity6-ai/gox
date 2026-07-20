package iobytez

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	ctx := context.Background()

	t.Run("read all", func(t *testing.T) {
		opts := &Options{}
		data := "hello world"
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		if string(opts.Out) != data {
			t.Errorf("expected %q, got %q", data, string(opts.Out))
		}
	})

	t.Run("with initial data", func(t *testing.T) {
		initial := "preexisting "
		opts := &Options{
			Out: []byte(initial),
		}
		data := "hello world"
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		expected := initial + data
		if string(opts.Out) != expected {
			t.Errorf("expected %q, got %q", expected, string(opts.Out))
		}
	})

	t.Run("with min", func(t *testing.T) {
		opts := &Options{Min: 5}
		data := "hello"
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		if string(opts.Out) != data {
			t.Errorf("expected %q, got %q", data, string(opts.Out))
		}
	})

	t.Run("min not met", func(t *testing.T) {
		opts := &Options{Min: 10}
		data := "hello"
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != io.ErrUnexpectedEOF {
			t.Errorf("expected io.ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("with max", func(t *testing.T) {
		opts := &Options{Max: 5}
		data := "hello world"
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		if string(opts.Out) != "hello" {
			t.Errorf("expected %q, got %q", "hello", string(opts.Out))
		}
	})

	t.Run("with min and max", func(t *testing.T) {
		opts := &Options{Min: 5, Max: 10}
		data := "hello world this is too long"
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		if string(opts.Out) != "hello worl" {
			t.Errorf("expected %q, got %q", "hello worl", string(opts.Out))
		}
	})

	t.Run("empty reader returns EOF", func(t *testing.T) {
		opts := &Options{}
		data := ""
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != io.EOF {
			t.Errorf("expected io.EOF, got %v", err)
		}
		if len(opts.Out) != 0 {
			t.Errorf("expected empty buffer, got %q", string(opts.Out))
		}
	})

	t.Run("read in chunks until EOF", func(t *testing.T) {
		fullData := "abcdefghijklmnopqrstuvwxyz"
		r := strings.NewReader(fullData)
		var totalReadBytes []byte

		// First chunk
		opts1 := &Options{Max: 10}
		err := Read(ctx, r, opts1)
		if err != nil {
			t.Fatalf("unexpected error on first read: %v", err)
		}
		expected1 := "abcdefghij"
		if string(opts1.Out) != expected1 {
			t.Errorf("expected first chunk %q, got %q", expected1, string(opts1.Out))
		}
		totalReadBytes = append(totalReadBytes, opts1.Out...)

		// Second chunk
		opts2 := &Options{Max: 10}
		err = Read(ctx, r, opts2)
		if err != nil {
			t.Fatalf("unexpected error on second read: %v", err)
		}
		expected2 := "klmnopqrst"
		if string(opts2.Out) != expected2 {
			t.Errorf("expected second chunk %q, got %q", expected2, string(opts2.Out))
		}
		totalReadBytes = append(totalReadBytes, opts2.Out...)

		// Third chunk (remaining data)
		opts3 := &Options{Max: 10}
		err = Read(ctx, r, opts3)
		if err != nil {
			t.Fatalf("unexpected error on third read: %v", err)
		}
		expected3 := "uvwxyz"
		if string(opts3.Out) != expected3 {
			t.Errorf("expected third chunk %q, got %q", expected3, string(opts3.Out))
		}
		totalReadBytes = append(totalReadBytes, opts3.Out...)

		// Fourth read (should be EOF)
		opts4 := &Options{Max: 10}
		err = Read(ctx, r, opts4)
		if err != io.EOF {
			t.Errorf("expected io.EOF on fourth read, got %v", err)
		}
		if len(opts4.Out) != 0 {
			t.Errorf("expected empty buffer on fourth read, got %q", string(opts4.Out))
		}
		
		// Verify total data read
		if string(totalReadBytes) != fullData {
			t.Errorf("expected total data %q, got %q", fullData, string(totalReadBytes))
		}
	})

	t.Run("empty reader with min", func(t *testing.T) {
		opts := &Options{Min: 1}
		data := ""
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != io.ErrUnexpectedEOF {
			t.Errorf("expected io.ErrUnexpectedEOF, got %v", err)
		}
	})
}
