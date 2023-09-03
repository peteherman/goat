package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Playbook struct {
	Name  string            `yaml:name`
	Hosts []string          `yaml:hosts`
	Vars  map[string]string `yaml:vars,omit=empty`
	Tasks []CommandTask     `yaml:tasks`
}

type CommandTask struct {
	Name string `yaml:name`
	Cmd  string `yaml:cmd`
}

const (
	NotInitiatedConnection = iota
	SuccessfulConnection
	FailedConnection
)

type PlaybookResult map[string]map[string]TaskResult

func PlaybookFromFilepath(filepath string) (Playbook, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return Playbook{}, err
	}
	return playbookFromContents(contents)
}

func playbookFromContents(contents []byte) (Playbook, error) {
	playbook := Playbook{
		Hosts: make([]string, 0),
		Vars:  make(map[string]string),
		Tasks: make([]CommandTask, 0),
	}
	err := yaml.Unmarshal(contents, &playbook)
	if err != nil {
		return Playbook{}, err
	}
	return playbook, nil
}

type executingHost struct {
	Host *Host
	conn Connection
}

type Connection interface {
	Connect(*Host) error
	Run(string) (TaskResult, error)
	Status() int
	SetConnectionError(error)
}

type TaskResult interface {
	Stdout() string
	StdoutBytes() []byte
	Stderr() string
	StderrBytes() []byte
	ReturnCode() int
}



func (p Playbook) Execute(inventory Inventory) PlaybookResult {

	hosts := inventory.ExecutionHosts(p.Hosts)
	executionHosts := make(map[string]executingHost, len(hosts))
	for _, host := range hosts {
		hostAddress := fmt.Sprintf("%v", host)
		executionHosts[hostAddress] = executingHost{
			Host: host,
			conn: &SSHConnection{},
		}
	}
	result := make(PlaybookResult, 0)
	for _, task := range p.Tasks {
		result[task.Name] = make(map[string]TaskResult, 0)
		for _, host := range hosts {
			executionHost := executionHosts[fmt.Sprintf("%v", host)]
			if status := executionHost.conn.Status(); status == FailedConnection {
				continue
			} else if status == NotInitiatedConnection {
				if err := executionHost.conn.Connect(executionHost.Host); err != nil {
					fmt.Printf("Error connecting to host: %v - %v\n", host.name, err)
					executionHost.conn.SetConnectionError(err)
					continue
				}
			}
			cmdResult, err := executionHost.conn.Run(task.Cmd)
			result[task.Name][host.name] = cmdResult
			fmt.Printf("Got error: %v\n", err)
		}
	}
	return result
}
