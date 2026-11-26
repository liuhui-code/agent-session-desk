package main

import (
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/document"
	"os"
)

func printAttentionDocument(sink *fixtureSink, reference string) error {
	if reference == "" {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := document.Resolve(reference, cwd); err != nil {
		return err
	}
	_, err = fmt.Fprintf(sink, "Review document: \"%s\"\n", reference)
	return err
}
