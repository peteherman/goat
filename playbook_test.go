package main

import (
	"os"
	"testing"
	"path/filepath"
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
		
}
