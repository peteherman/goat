package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var simpleInventory = []byte(`
all:
  hosts:
    ssh1:
      ssh_port: 2222
      username: test
      password: test
`)
var singleNestedInventory = []byte(`
all:
  hosts:
    ssh1:
      vars:
        ssh_port: 2222
        username: test1
        password: test1
`)
var doublyNestedInventory = []byte(`
all:
  hosts:
    ssh1:
      vars:
        ssh_port: 2222
        username: test1
        password: test1
  children:
     ssh3:
       vars:
         ssh_port: 2222
         username: test3
         password: test3
`)
var triplyNestedInventory = []byte(`
all:
  hosts:
    ssh1:
      vars:
        ssh_port: 2222
        username: test
        password: test
        outer_only: yes
    ssh2:
      vars:
        ssh_port: 2222
        username: test2
        password: test2
    ssh4:
  children:
     ssh2:
       hosts:
         ssh3:
     inner:
       hosts:
         ssh1:
           vars:
             ssh_port: 2222
             username: test_inner
             password: test_inner
`)

func TestInventoryFromFileWhenNotExist(t *testing.T) {
	bsFilepath := "idontexist"
	_, err := InventoryFromFilepath(bsFilepath)
	if err == nil {
		t.Fatalf("Expected an error for this, creating inventory from non-existant filepath!\n")
	}
}

func TestInventoryFromSimpleFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, simpleInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}
	if len(inventory.All.Hosts) <= 0 {
		t.Fatalf("Expected at least one host in inventory.All.Hosts!\n")
	}
}

func TestInventoryWithSingleNestedHostGroups(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, singleNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}
	if len(inventory.All.Hosts) <= 0 {
		t.Fatalf("Expected at least one host in inventory.All.Hosts!\n")
	}
}

func TestInventoryWithDoublyNestedHostGroups(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, doublyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}
	if len(inventory.All.Hosts) <= 0 {
		t.Fatalf("Expected at least one host in inventory.All.Hosts!\n")
	}
	if len(inventory.All.Children) <= 0 {
		t.Fatalf("Expected at least one hostgroup in inventory.All.Children!\n")
	}
}

func TestInventoryWithTriplyNestedHostGroups(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}
	if len(inventory.All.Hosts) <= 0 {
		t.Fatalf("Expected at least one host in inventory.All.Hosts!\n")
	}
	if len(inventory.All.Children) <= 0 {
		t.Fatalf("Expected at least one hostgroup in inventory.All.Children!\n")
	}
	if len(inventory.All.Children["inner"].Hosts) <= 0 {
		t.Fatalf("Expected at least one hostgroup in inventory.All.Children[\"inner\"]!\n")
	}
}

func TestGatherWithNestedHostGroup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}
	if len(inventory.All.Children["inner"].Hosts) <= 0 {
		t.Fatalf("Inventory wasn't properly initialized for this test: %v\n", err)
	}

	_, err = inventory.gatherHosts("ssh1")
	if err != nil {
		fmt.Printf("inventory.All.Hosts: %+v\n", inventory.All.Hosts)
		t.Fatalf("Didn't find host 'ssh1' but should've: %v\n", err)
	}
}

func TestGatherHostNotFoundErr(t *testing.T) {
	inventory := Inventory{}
	_, err := inventory.gatherHosts("not there")
	if err == nil {
		t.Fatalf("Expected HostNotFound error, got nil\n")
	} else {
		invErr, ok := err.(*InventoryError)
		if ok {
			if !invErr.HostNotFound() {
				t.Fatalf("Error should've been HostNotFound\n")
			}
		}
	}
}

func TestHostGroupGatherHostNotFoundErr(t *testing.T) {
	hg := HostGroup{}
	_, err := hg.gatherHosts("test")
	if err == nil {
		t.Fatalf("Expected HostNotFound error, got nil\n")
	} else {
		invErr, ok := err.(*InventoryError)
		if ok {
			if !invErr.HostNotFound() {
				t.Fatalf("Error should've been HostNotFound: %v\n", err)
			}
		}
	}	
}

func TestGatherWithNestedHost(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}
	host, err := inventory.gatherHosts("ssh1")
	if err != nil {
		t.Fatalf("Received error when gathering existing host: %v\n", err)
	}
	if host.Vars["username"] != "test_inner" {
		t.Fatalf("Layered username didn't overwrite upper username\n")
	}
	if _, keyExists := host.Vars["outer_only"]; !keyExists {
		t.Fatalf("Host outer only variable wasn't written\n")
	}
}

func TestExecutionHostsWithHostnames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}

	hosts, err := inventory.ExecutionHosts([]string{"ssh2"})
	if err != nil {
		t.Fatalf("Received error when getting execution hosts with known host: %v\n", err)
	}
	if len(hosts) <= 0 {
		t.Fatalf("Didn't receive any hosts when trying to get Execution Hosts\n")
	}
	if len(hosts[0].Vars) <= 0 {
		t.Fatalf("Host didn't contain any vars\n")
	}
}

func TestExecutionHostsWithHostGroupName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}

	hosts, err := inventory.ExecutionHosts([]string{"inner"})
	if err != nil {
		t.Fatalf("Received error when getting execution hosts with known host group : %v\n", err)
	}
	if len(hosts) <= 0 {
		t.Fatalf("Didn't receive any hosts when trying to get Execution Hosts\n")
	}
	if len(hosts[0].Vars) <= 0 {
		t.Fatalf("Host didn't contain any vars\n")
	}
}

func TestExecutionHostsWithHostNameAndGroupNameMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}

	hosts, err := inventory.ExecutionHosts([]string{"ssh2"})
	if err != nil {
		t.Fatalf("Received error when getting execution hosts with known host group : %v\n", err)
	}
	if len(hosts) <= 1 {
		t.Fatalf("Didn't receive enough hosts when trying to get Execution Hosts\n")
	}
	if len(hosts[0].Vars) <= 0 {
		t.Fatalf("Host didn't contain any vars\n")
	}
}

func TestExecutionHostsWithMultipleHostnames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}

	hosts, err := inventory.ExecutionHosts([]string{"ssh1", "ssh4"})
	if err != nil {
		t.Fatalf("Received error when getting execution hosts with known host group : %v\n", err)
	}
	if len(hosts) <= 1 {
		t.Fatalf("Didn't receive enough hosts when trying to get Execution Hosts\n")
	}
	if len(hosts[0].Vars) <= 0 {
		t.Fatalf("Host didn't contain any vars\n")
	}
}

func TesetExecutionHostsWithAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, triplyNestedInventory, 0666)
	if err != nil {
		t.Fatalf("Error when creating test directory\n")
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		t.Fatalf("Received error on Inventory Parse from file: %v\n", err)
	}

	hosts, err := inventory.ExecutionHosts([]string{"all"})
	if err != nil {
		t.Fatalf("Received error when getting execution hosts with known host group : %v\n", err)
	}
	if len(hosts) <= 1 {
		t.Fatalf("Didn't receive enough hosts when trying to get Execution Hosts\n")
	}
	if len(hosts[0].Vars) <= 0 {
		t.Fatalf("Host didn't contain any vars\n")
	}	
}
