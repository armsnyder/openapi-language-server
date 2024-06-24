package e2etest_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	const testdataDir = "testdata"

	files, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			t.Run(file.Name(), func(t *testing.T) {
				runWithTestData(t, filepath.Join(testdataDir, file.Name()))
			})
		}
	}
}

func runWithTestData(t *testing.T, testDataDir string) {
	inputData, err := os.ReadFile(testDataDir + `/input.jsonrpc`)
	if err != nil {
		t.Fatalf("Failed to read input data: %v", err)
	}

	expectedOutputData, err := os.ReadFile(testDataDir + "/output.jsonrpc")
	if err != nil {
		t.Fatalf("Failed to read expected output data: %v", err)
	}

	output := runWithInput(t, bytes.NewReader(inputData))

	format := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, "\r", "␍"), "\n", "␊")
	}

	if string(expectedOutputData) != output {
		t.Errorf("Output did not match expectation.\n\ngot\n%s\n\nwant\n%s\n", format(output), format(string(expectedOutputData)))
	}
}

func runWithInput(t *testing.T, stdin io.Reader) string {
	var buf bytes.Buffer

	run(t, func(cmd *exec.Cmd) {
		cmd.Stdin = stdin
		cmd.Stdout = &buf
		cmd.Stderr = &testlogger{t: t, prefix: "server: "}
	})

	return buf.String()
}

func run(t *testing.T, configureCommand func(cmd *exec.Cmd)) {
	const binName = "openapi-language-server"

	cmd := exec.Command(binName)

	t.Cleanup(func() {
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				t.Fatalf("Warning: %s process may still be running: %v", binName, err)
			}
		}
	})

	configureCommand(cmd)

	if err := cmd.Start(); err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			if os.Getenv("CI") == "true" {
				t.Fatalf("End-to-end tests require %s to be installed. Error: %v", binName, err)
			}
			t.Skipf("End-to-end tests require %s to be installed. Error: %v", binName, err)
		}
		t.Fatal(err)
	}

	doneCh := make(chan error, 1)

	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for server to exit")
	}
}

type testlogger struct {
	t      *testing.T
	prefix string
}

// Write implements io.Writer.
func (t *testlogger) Write(p []byte) (n int, err error) {
	t.t.Log(t.prefix + string(p))
	return len(p), nil
}

var _ io.Writer = (*testlogger)(nil)
