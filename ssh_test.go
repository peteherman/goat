package main

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"testing"
)

func TestConnectionStatusNotInit(t *testing.T) {
	connection := SSHConnection{}
	if status := connection.Status(); status != NotInitiatedConnection {
		t.Fatalf("Expected ssh connection status: 'NotInitiatedConnection' got another: %v\n", status)
	}
}

func TestSuccessfulConnection(t *testing.T) {
	connection := SSHConnection{
		Client:    &ssh.Client{},
		connError: nil,
	}
	if status := connection.Status(); status != SuccessfulConnection {
		t.Fatalf("Expected ssh connection status: 'SuccessfulConnection' got another: %v\n", status)
	}
}

func TestFailedConnection(t *testing.T) {
	connection := SSHConnection{
		Client:    &ssh.Client{},
		connError: errors.New("not nothing here"),
	}
	if status := connection.Status(); status != FailedConnection {
		t.Fatalf("Expected ssh connection status: 'FailedConnection' got another: %v\n", status)
	}

	connection = SSHConnection{
		Client: nil,
		connError: errors.New("not nothing again"),
	}
	if status := connection.Status(); status != FailedConnection {
		t.Fatalf("Expected ssh connection status: 'FailedConnection' got another: %v\n", status)
	}	
}

func TestConfigFromHostErrorNoUsernameWhenNoUsername(t *testing.T) {
	host := Host{
		Vars: make(map[string]string, 0),
		name: "test host",
	}
	_, err := configFromHost(&host)
	if err == nil {
		t.Fatalf("Expected error here for host not having a 'username' specified.")
	}	
}

func TestConfigFromHostErrorNoCreds(t *testing.T) {
	host := Host{
		Vars: make(map[string]string, 0),
		name: "test host",
	}
	host.Vars["username"] = "test"
	_, err := configFromHost(&host)
	if err == nil {
		t.Fatalf("Expected error here for host not having any auth creds (password or key) specified.")
	}	
}

