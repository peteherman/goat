package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	HostNotFoundErrorCode = iota
)

type InventoryError struct {
	errorCode int
	Err       error
}

func (i *InventoryError) HostNotFound() bool {
	return i.errorCode == HostNotFoundErrorCode
}
func (i *InventoryError) Error() string {
	return i.Err.Error()
}

type Host struct {
	Vars map[string]string `yaml:vars,omit=empty`
}

type HostGroup struct {
	Hosts    map[string]Host      `yaml:hosts,omit=empty`
	Children map[string]HostGroup `yaml:children,omit=empty`
	Vars     map[string]string    `yaml:vars,omit=empty`
}

type Inventory struct {
	All      HostGroup            `yaml:all,omit=empty`
	Vars     map[string]string    `yaml:vars,omit=empty`
}

func InventoryFromFilepath(filepath string) (Inventory, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return Inventory{}, err
	}

	return inventoryFromFileContents(contents)
}

func inventoryFromFileContents(contents []byte) (Inventory, error) {
	inventory := Inventory{
		All: HostGroup{
			Hosts:    make(map[string]Host, 0),
			Children: make(map[string]HostGroup, 0),
			Vars:     make(map[string]string, 0),
		},
		Vars:     make(map[string]string, 0),
	}
	err := yaml.Unmarshal(contents, &inventory)
	if err != nil {
		return Inventory{}, err
	}
	return inventory, nil
}

func (i Inventory) gather(hostname string) (*Host, error) {
	return i.All.gather(hostname)
}

// Need to perform depth first search through host group's children.
// The variables in the lower hostgroups will take precedence over the variables
// in the upper, more generic hostgroups
func (g HostGroup) gather(hostname string) (*Host, error) {
	host := &Host{
		Vars: make(map[string]string, 0),
	}
	for hostKey, existingHost := range g.Hosts {
		if hostKey == hostname {
			for variableKey, value := range existingHost.Vars {
				if _, keyExists := host.Vars[variableKey]; !keyExists {
					host.Vars[variableKey] = value
				}
			}
		}
	}
	for _, hostgroup := range g.Children {
		bubbledHost, err := hostgroup.gather(hostname)
		if err != nil {
			invErr, ok := err.(*InventoryError)
			if ok {
				if invErr.HostNotFound() {
					continue
				}
			}
			return nil, err
		}
		for variableKey, value := range bubbledHost.Vars {
			host.Vars[variableKey] = value
		}

	}

	if len(host.Vars) <= 0 {
		return nil, &InventoryError{
			errorCode: HostNotFoundErrorCode,
			Err:       errors.New(fmt.Sprintf("Unable to locate host: %v\n", hostname)),
		}
	}
	return host, nil
}
