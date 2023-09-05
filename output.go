package main

import (
	"bufio"
	"fmt"
	"strings"
)

type OutputFormatter interface {
	Output(string, string, TaskResult) string
}

type StdoutFormatter struct{}

func (s StdoutFormatter) Output(taskName, hostname string, result TaskResult) string {

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("task: %v\n", taskName))
	sb.WriteString(fmt.Sprintf("\thost: %v\n", hostname))

	taskErr := result.Error()
	if taskErr != nil {
		sb.WriteString(fmt.Sprintf("\t\terror: %v\n", taskErr.Error()))
		stdout := result.Stdout()
		scanner := bufio.NewScanner(strings.NewReader(stdout))
		sb.WriteString(fmt.Sprintf("\t\tstdout:\n"))
		for scanner.Scan() {
			sb.WriteString(fmt.Sprintf("\t\t\t%v\n", strings.TrimSpace(scanner.Text())))
		}
		stderr := result.Stderr()		
		scanner = bufio.NewScanner(strings.NewReader(stderr))
		sb.WriteString(fmt.Sprintf("\t\tstderr:\n"))
		for scanner.Scan() {
			sb.WriteString(fmt.Sprintf("\t\t\t%v\n", strings.TrimSpace(scanner.Text())))
		}
		return sb.String()
	}

	stdout := result.Stdout()
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	sb.WriteString(fmt.Sprintf("\t\tstdout:\n"))
	for scanner.Scan() {
		sb.WriteString(fmt.Sprintf("\t\t\t%v\n", strings.TrimSpace(scanner.Text())))
	}
	stderr := result.Stderr()	
	scanner = bufio.NewScanner(strings.NewReader(stderr))
	sb.WriteString(fmt.Sprintf("\t\tstderr:\n"))
	for scanner.Scan() {
		sb.WriteString(fmt.Sprintf("\t\t\t%v\n", strings.TrimSpace(scanner.Text())))
	}
	return sb.String()

}
