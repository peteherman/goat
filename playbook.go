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

type SSHConnection struct {
	Client    *ssh.Client
	connError error
}

func (s *SSHConnection) SetConnectionError(err error) {
	s.connError = err
}

func (s *SSHConnection) Status() int {
	if s.connError == nil && s.Client == nil {
		return NotInitiatedConnection
	}
	if s.connError == nil && s.Client != nil {
		return SuccessfulConnection
	}
	return FailedConnection
}

func (s *SSHConnection) Run(command string) (TaskResult, error) {
	if status := s.Status(); status != SuccessfulConnection {
		if status == FailedConnection {
			return SSHCommandResult{}, errors.New("Connection failed")
		}
		return SSHCommandResult{}, errors.New("Connection not initiated")
	}
	session, err := s.Client.NewSession()
	if err != nil {
		s.connError = err
		return SSHCommandResult{}, errors.New("Unable to create session on host\n")
	}
	defer session.Close()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(command)

	return SSHCommandResult{
		stdoutBuffer: stdout,
		stderrBuffer: stderr,
		returnCode:   0,
	}, nil
}

func (s *SSHConnection) Connect(host *Host) error {
	username, keyExists := host.Vars["username"]
	if !keyExists {
		return errors.New(fmt.Sprintf("Cannot connect to host %v, no username provided\n",
			host.name))
	}
	password, keyExists := host.Vars["password"]
	if !keyExists {
		return errors.New(fmt.Sprintf("Cannot connect to host %v, no password provided\n",
			host.name))
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	address := host.name
	if newAddress, keyExists := host.Vars["address"]; keyExists {
		address = newAddress
	}

	port := "22"
	if specifiedPort, keyExists := host.Vars["port"]; keyExists {
		port = specifiedPort
	}

	uri := fmt.Sprintf("%v:%v", address, port)

	client, err := ssh.Dial("tcp", uri, config)
	if err != nil {
		return errors.New(fmt.Sprintf("Error when connecting to host: %v\n", err))
	}
	s.Client = client

	return nil
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
