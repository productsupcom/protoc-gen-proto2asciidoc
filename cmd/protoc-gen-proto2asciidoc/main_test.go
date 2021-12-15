package main

import (
	"fmt"
	"os"
	"testing"
)

func Test_process(t *testing.T) {
	file, err := os.Open("../../protowire")
	if err != nil {
		t.Logf("did you run protoc before?")
		t.Fatalf("could not open file: %v", err)
	}

	fmt.Fprintf(os.Stdout, "\n%s\n", process(file))
}
