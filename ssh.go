package main

import (
	"bytes"
	"errors"
	"fmt"
	"time"
	"golang.org/x/crypto/ssh"
)

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

func (s *SSHConnection) Run(command string) TaskResult {
	if status := s.Status(); status != SuccessfulConnection {
		if status == FailedConnection {
			return SSHCommandResult{
				err: errors.New("Connection failed"),
			}
		}
		return SSHCommandResult{
			err: errors.New("Connection not initiated"),
		}
	}
	session, err := s.Client.NewSession()
	if err != nil {
		s.connError = err
		return SSHCommandResult{
			err: errors.New("Unable to create session on host\n"),
		}
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
		err:          err,
	}
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
		Timeout: time.Second * 5,
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
	err          error
}

func (s SSHCommandResult) StdoutBytes() []byte {
	return s.stdoutBuffer.Bytes()
}

func (s SSHCommandResult) StderrBytes() []byte {
	return s.stderrBuffer.Bytes()
}

func (s SSHCommandResult) Stdout() string {
	return s.stdoutBuffer.String()
}

func (s SSHCommandResult) Stderr() string {
	return s.stderrBuffer.String()
}

func (s SSHCommandResult) Error() error {
	return s.err
}
