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

	t.Run("empty reader", func(t *testing.T) {
		opts := &Options{}
		data := ""
		r := strings.NewReader(data)
		err := Read(ctx, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		if string(opts.Out) != "" {
			t.Errorf("expected empty string, got %q", string(opts.Out))
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
