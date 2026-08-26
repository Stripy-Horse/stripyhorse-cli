package main

import (
	"fmt"
	"io"
	"os"

	stripyhorse "github.com/Stripy-Horse/stripyhorse-go"
)

// readInput reads from the named file, or from stdin when the name is empty
// or "-", so every command that takes ZPL is pipe-friendly.
func readInput(name string) ([]byte, error) {
	if name == "" || name == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(name)
}

// writeOut writes to the named file, or to stdout when the name is empty.
func writeOut(name string, data []byte) error {
	if name == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	if err := os.WriteFile(name, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", name, len(data))
	return nil
}

// apiError unwraps the SDK's error to surface the server's error body, which
// carries the actual reason (invalid_zpl, quota_exceeded, …).
func apiError(err error) error {
	var ge *stripyhorse.GenericOpenAPIError
	if e, ok := err.(*stripyhorse.GenericOpenAPIError); ok {
		ge = e
	}
	if ge != nil && len(ge.Body()) > 0 {
		return fmt.Errorf("%s: %s", ge.Error(), string(ge.Body()))
	}
	return err
}
