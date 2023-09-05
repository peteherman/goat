package main

import (
	"strings"
	"fmt"
)

type OutputFormatter interface {
	Output(string, string, TaskResult) string
}

type StdoutFormatter struct {}

func (s StdoutFormatter) Output(taskName, hostname string, result TaskResult) string {

	var sb strings.Builder	
	sb.WriteString(fmt.Sprintf("task: %v\n", taskName))
	sb.WriteString(fmt.Sprintf("\thost: %v\n", hostname))

	taskErr := result.Error()
	if taskErr != nil {
		sb.WriteString(fmt.Sprintf("\t\terror: %v\n", taskErr.Error()))
		return sb.String()
	}
	
	stdout := result.Stdout()
	sb.WriteString(fmt.Sprintf("\t\tstdout: %v\n", strings.TrimSpace(stdout)))
	stderr := result.Stderr()
	sb.WriteString(fmt.Sprintf("\t\tstderr: %v\n", strings.TrimSpace(stderr)))
	return sb.String()
	
}


