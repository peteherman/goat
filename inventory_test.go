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

var executionHostsNoGroup = []byte(`
all:
  hosts:
    ssh1:
      vars:
        username: inside
  vars:
    username: outside
`)

var executionHostsOneGroup = []byte(`
all:
  children:
    group:
      hosts:
        ssh2:
        ssh1:
          vars: 
            username: inside
      vars:
        username: outside
`)

var executionHostsTwoSubs = []byte(`
all:
  children:
    dmz:
      children:
        linux:
          hosts: 
            ssh1:
              vars:
                username: inside
          vars: 
            username: middle
    bank: 
      children:
        linux:
          hosts: 
            ssh2:
  vars:
    username: outer
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

func TestExecutionHostsWithNoGroupMatches(t *testing.T) {
	inventory, err := buildInventory(t, executionHostsNoGroup)
	if err != nil {
		t.Fatalf("Failed to initialize inventory for test! %v\n", err)
	}
	hosts := inventory.ExecutionHosts([]string{"ssh1"})
	if len(hosts) <= 0 {
		t.Fatalf("Didn't receive enough hosts when trying to get Execution Hosts\n")
	}
	if len(hosts[0].Vars) <= 0 {
		t.Fatalf("Host didn't contain any vars\n")
	}
	if hosts[0].Vars["username"] != "inside" {
		t.Fatalf("Host username is incorrect, should be 'inside' got: %v\n",
			hosts[0].Vars["username"])
	}
}

func TestExecutionHostsWithOneGroupMatch(t *testing.T) {
	inventory, err := buildInventory(t, executionHostsOneGroup)
	if err != nil {
		t.Fatalf("Failed to initialize inventory for test! %v\n", err)
	}
	hosts := inventory.ExecutionHosts([]string{"group"})
	if len(hosts) < 2 {
		fmt.Printf("Inventory: %+v\n", inventory)
		t.Fatalf("Didn't receive enough hosts when trying to get Execution Hosts - %+v\n", hosts)
	}
	if len(hosts[0].Vars) <= 0 {
		t.Fatalf("Host didn't contain any vars\n")
	}
	for _, host := range hosts {
		if host.name == "ssh1" && host.Vars["username"] != "inside" {
			t.Fatalf("Host username is incorrect, should be 'inside' got: %v, %+v\n",
				hosts[0].Vars["username"], *host)

		}
	}
}

func TestExecutionHostsWithTwoSubMatch(t *testing.T) {
	inventory, err := buildInventory(t, executionHostsTwoSubs)
	if err != nil {
		t.Fatalf("Failed to initialize inventory for test! %v\n", err)
	}
	hosts := inventory.ExecutionHosts([]string{"linux"})
	if len(hosts) < 2 {
		fmt.Printf("Inventory: %+v\n", inventory)
		t.Fatalf("Didn't receive enough hosts when trying to get Execution Hosts - %+v\n", hosts)
	}

	for _, host := range hosts {
		if host.name == "ssh1" && host.Vars["username"] != "inside" {
			t.Fatalf("Host username incorrect, expected 'inside': %+v\n", *host)
		}
		if host.name == "ssh2" && host.Vars["username"] != "outer" {
			t.Fatalf("Host username incorrect, expected 'outer': %+v\n", *host)
		}
	}
}

func buildInventory(t *testing.T, contents []byte) (Inventory, error) {
	tmpDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		return Inventory{}, err
	}
	tmpFilename := filepath.Join(tmpDir, fmt.Sprintf("%v.yaml", t.Name()))
	err = os.WriteFile(tmpFilename, contents, 0666)
	if err != nil {
		return Inventory{}, err
	}
	inventory, err := InventoryFromFilepath(tmpFilename)
	if err != nil {
		return Inventory{}, err
	}
	return inventory, nil
}
