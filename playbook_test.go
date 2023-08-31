package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"os"
	"path/filepath"
	"testing"
)

var shortPlaybookContents = []byte(`
name: Short Playbook
hosts: 
  - all
vars: 
  sample_var: here
tasks:
  - name: task one
    cmd: whoami
`)

func TestPlaybookFromNonExistentFile(t *testing.T) {
	_, err := PlaybookFromFilepath("nonexistent")
	if err == nil {
		t.Fatalf("Should've received an error when creating file from nonexistent path")
	}
}

func TestPlaybookFromShortPlaybook(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatalf("Unable to initialize temp directory for this test!\n")
	}

	tmpFilepath := filepath.Join(tmpDir, "playbook.yaml")
	err = os.WriteFile(tmpFilepath, shortPlaybookContents, 0666)
	if err != nil {
		t.Fatalf("Unable to write temp playbook file for this test: %v\n", t.Name())
	}

	playbook, err := PlaybookFromFilepath(tmpFilepath)
	if err != nil {
		t.Fatalf("Received error when parsing playbook: %v\n", err)
	}
	if playbook.Name != "Short Playbook" {
		t.Fatalf("Playbook didn't contain correct name")
	}
	if len(playbook.Hosts) <= 0 {
		t.Fatalf("Playbook didn't contain any hosts")
	}
	if len(playbook.Hosts) <= 0 {
		t.Fatalf("Playbook didn't contain any hosts")
	}
	if len(playbook.Vars) <= 0 {
		t.Fatalf("Expected at least one Var in shortPlaybook\n")
	}
	if len(playbook.Tasks) <= 0 {
		t.Fatalf("Expected at least one Task in shortPlaybook\n")
	}
	expectedTaskName := "task one"
	taskOne := playbook.Tasks[0]
	if taskOne.Name != expectedTaskName {
		t.Fatalf("Expected %v taskname in tasks: %v\n", expectedTaskName, playbook)
	}
	if taskOne.Cmd != "whoami" {
		t.Fatalf("Expected whoami command in tasks: %v\n", playbook)
	}	
}

func TestPlaybookExecuteShortResults(t *testing.T) {
	hostGroup := HostGroup{
		Hosts: make(map[string]Host, 1),
	}

	dockerSSHContainer1 := dockerSSH1Host()
	hostGroup.Hosts["ssh1"] = dockerSSHContainer1

	inventory := Inventory{
		All: hostGroup,
	}

	playbook := createTestPlaybook(t, shortPlaybookContents)

	testContainersRunning(t, []string{"ssh1"})
	
	playbookResults := playbook.Execute(inventory)
	if len(playbookResults) <= 0 {
		t.Fatalf("Expected one task result for small playbook!\n")
	}
	expectedTaskName := "task one"
	if _, keyExists := playbookResults[expectedTaskName]; !keyExists {
		t.Fatalf("Expected %v in playbookResults: %v\n", expectedTaskName, playbookResults)
	}

	expectedHostname := "ssh1"
	if _, keyExists := playbookResults[expectedTaskName][expectedHostname]; !keyExists {
		t.Fatalf("Expected %v in playbookResults: %v\n", expectedHostname, playbookResults)
	}	
}

func dockerSSH1Host() Host {
	dockerSSHContainer1 := Host{
		Vars: make(map[string]string, 4),
	}
	dockerSSHContainer1.Vars["address"] = "localhost"
	dockerSSHContainer1.Vars["port"] = "2221"
	dockerSSHContainer1.Vars["username"] = "test"
	dockerSSHContainer1.Vars["password"] = "test"
	return dockerSSHContainer1
}

func createTestPlaybook(t *testing.T, contents []byte) Playbook {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatalf("Unable to initialize temp directory for this test!\n")
	}

	tmpFilepath := filepath.Join(tmpDir, "playbook.yaml")
	err = os.WriteFile(tmpFilepath, contents, 0666)
	if err != nil {
		t.Fatalf("Unable to write temp playbook file for this test: %v\n", t.Name())
	}

	playbook, err := PlaybookFromFilepath(tmpFilepath)
	if err != nil {
		t.Fatalf("Received error when parsing playbook: %v\n", err)
	}
	return playbook
}

func testContainersRunning(t *testing.T, containerNames []string) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.41"), client.FromEnv)
	if err != nil {
		t.Fatalf("Error initializing docker client: %v\n", err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		t.Fatalf("Error fetching containers: %v\n", err)
	}
	containerFound := false
	for _, searchName := range containerNames {
		containerFound = false
		for _, container := range containers {
			for _, name := range container.Names {
				testName := fmt.Sprintf("/%v", searchName)
				if name == testName && container.State == "running" {
					containerFound = true
					break
				}
			}
			if containerFound {
				break
			}
		}
		if !containerFound {
			t.Fatalf("container %v was not found or running!\n", searchName)
		}
	}
}
