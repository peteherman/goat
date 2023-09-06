package main

import (
	"fmt"
	"strings"
	"testing"
	"errors"
)

type TestResult struct {
	stdout []byte
	stderr []byte
	err    error
}

func (t TestResult) Stdout() string {
	return string(t.stdout)
}
func (t TestResult) StdoutBytes() []byte {
	return t.stdout
}

func (t TestResult) Stderr() string {
	return string(t.stderr)
}

func (t TestResult) StderrBytes() []byte {
	return t.stderr
}

func (t TestResult) Error() error {
	return t.err
}

func TestStdoutFormatterNoError(t *testing.T) {
	res := TestResult{
		stdout: []byte(`stdout here`),
		stderr: []byte(`stderr here`),
		err:    nil,
	}
	taskname := "test taskname"
	hostname := "test hostname"

	formatter := StdoutFormatter{}
	s := formatter.Output(taskname, hostname, res)

	outputSplit := strings.Split(s, "\n")
	if l := len(outputSplit); l != 7 {
		t.Fatalf("Expected output to contain 7 lines. Got: %v\n", l)
	}

	expectedTaskline := fmt.Sprintf("task: %v", taskname)
	if outputSplit[0] != expectedTaskline {
		t.Fatalf("First line of output should contain 'task: <task name>', it did not: %v\n", outputSplit[0])
	}
	expectedHostline := fmt.Sprintf("host: %v", hostname)
	if hostline := strings.TrimSpace(outputSplit[1]); hostline != expectedHostline {
		t.Fatalf("Second line of output should contain 'host: <host name>', it did not: %v\n", hostline)		
	}
	expectedStdoutline := fmt.Sprintf("stdout:")
	if stdoutline := strings.TrimSpace(outputSplit[2]); stdoutline != expectedStdoutline {
		t.Fatalf("Third line of output should contain: 'stdout:', but it didn't: %v\n", stdoutline)
	}
	expectedStdoutline = fmt.Sprintf("stdout here")
	if stdoutline := strings.TrimSpace(outputSplit[3]); stdoutline != expectedStdoutline {
		t.Fatalf("Fourth line of output should contain: 'stdout:', but it didn't: %v\n", stdoutline)
	}
	expectedStderrline := fmt.Sprintf("stderr:")
	if stderrline := strings.TrimSpace(outputSplit[4]); stderrline != expectedStderrline {
		t.Fatalf("Fifth line of output should contain: 'stderr:', but it didn't: %v\n", stderrline)
	}
	expectedStderrline = fmt.Sprintf("stderr here")
	if stderrline := strings.TrimSpace(outputSplit[5]); stderrline != expectedStderrline {
		t.Fatalf("Sixth line of output should contain: 'stderr:', but it didn't: %v\n", stderrline)
	}		

}

func TestStdoutFormatterWithError(t *testing.T) {
	res := TestResult{
		stdout: []byte(`stdout here`),
		stderr: []byte(`stderr here`),
		err:    errors.New("Here's an error"),
	}
	taskname := "test taskname"
	hostname := "test hostname"

	formatter := StdoutFormatter{}
	s := formatter.Output(taskname, hostname, res)

	outputSplit := strings.Split(s, "\n")
	if l := len(outputSplit); l != 8 {
		t.Fatalf("Expected output to contain 8 lines. Got: %v\n", l)
	}

	expectedTaskline := fmt.Sprintf("task: %v", taskname)
	if outputSplit[0] != expectedTaskline {
		t.Fatalf("First line of output should contain 'task: <task name>', it did not: %v\n", outputSplit[0])
	}
	expectedHostline := fmt.Sprintf("host: %v", hostname)
	if hostline := strings.TrimSpace(outputSplit[1]); hostline != expectedHostline {
		t.Fatalf("Second line of output should contain 'host: <host name>', it did not: %v\n", hostline)		
	}
	expectedErrline := fmt.Sprintf("error: Here's an error")
	if errLine := strings.TrimSpace(outputSplit[2]); errLine != expectedErrline {
		t.Fatalf("Third line of output should contain: 'error: Here's an error' but it didn't: %v\n", errLine)
	}
	expectedStdoutline := fmt.Sprintf("stdout:")
	if stdoutline := strings.TrimSpace(outputSplit[3]); stdoutline != expectedStdoutline {
		t.Fatalf("Fourth line of output should contain: 'stdout:', but it didn't: %v\n", stdoutline)
	}
	expectedStdoutline = fmt.Sprintf("stdout here")
	if stdoutline := strings.TrimSpace(outputSplit[4]); stdoutline != expectedStdoutline {
		t.Fatalf("Fifth line of output should contain: 'stdout:', but it didn't: %v\n", stdoutline)
	}
	expectedStderrline := fmt.Sprintf("stderr:")
	if stderrline := strings.TrimSpace(outputSplit[5]); stderrline != expectedStderrline {
		t.Fatalf("Sixth line of output should contain: 'stderr:', but it didn't: %v\n", stderrline)
	}
	expectedStderrline = fmt.Sprintf("stderr here")
	if stderrline := strings.TrimSpace(outputSplit[6]); stderrline != expectedStderrline {
		t.Fatalf("Seventh line of output should contain: 'stderr:', but it didn't: %v\n", stderrline)
	}			
}
