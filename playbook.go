package main

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
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
	Host       *Host
	connStatus int
	conn       Connection
}

type Connection interface {
	Connect(*Host) error
	Run(string) (TaskResult, error)
}

type TaskResult interface {
	Stdout() string
	StdoutBytes() []byte
	Stderr() string
	StderrBytes() []byte
	ReturnCode() int
}

type SSHConnection struct {
	client *ssh.Client
}

func (s SSHConnection) Run(command string) (TaskResult, error) {
	return SSHCommandResult{}, nil
}

func (s SSHConnection) Connect(host *Host) error {
	return errors.New("Haven't implemented the ssh connection yet")
}

type SSHCommandResult struct {
	stdoutBuffer bytes.Buffer
	stderrBuffer bytes.Buffer
	returnCode   int
}

func (s SSHCommandResult) StdoutBytes() []byte {
	return s.stdoutBuffer.Bytes()
}

func (s SSHCommandResult) StderrBytes() []byte {
	return s.stderrBuffer.Bytes()
}

func (s SSHCommandResult) ReturnCode() int {
	return s.returnCode
}

func (s SSHCommandResult) Stdout() string {
	return s.stdoutBuffer.String()
}

func (s SSHCommandResult) Stderr() string {
	return s.stderrBuffer.String()
}

func (p Playbook) Execute(inventory Inventory) PlaybookResult {

	hosts := inventory.ExecutionHosts(p.Hosts)
	executionHosts := make(map[string]executingHost, len(hosts))
	for _, host := range hosts {
		hostAddress := fmt.Sprintf("%v", host)
		executionHosts[hostAddress] = executingHost{
			Host:       host,
			connStatus: NotInitiatedConnection,
			conn:       SSHConnection{},
		}
	}
	for _, task := range p.Tasks {
		for _, host := range hosts {
			executionHost := executionHosts[fmt.Sprintf("%v", host)]
			if executionHost.connStatus == FailedConnection {
				continue
			} else if executionHost.connStatus == NotInitiatedConnection {
				if err := executionHost.conn.Connect(executionHost.Host); err != nil {
					fmt.Printf("Error connecting to host: %v\n", host.name)
					executionHost.connStatus = FailedConnection
					continue
				}
			}
			executionHost.conn.Run(task.Cmd)
		}
	}
	return make(PlaybookResult, 0)
}
