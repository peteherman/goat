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
	All  HostGroup         `yaml:all,omit=empty`
	Vars map[string]string `yaml:vars,omit=empty`
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
		Vars: make(map[string]string, 0),
	}
	err := yaml.Unmarshal(contents, &inventory)
	if err != nil {
		return Inventory{}, err
	}
	return inventory, nil
}

func (i Inventory) gatherHosts(hostname string) (*Host, error) {
	return i.All.gatherHosts(hostname)
}

// Need to perform depth first search through host group's children.
// The variables in the lower hostgroups will take precedence over the variables
// in the upper, more generic hostgroups
func (g HostGroup) gatherHosts(hostname string) (*Host, error) {
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
		bubbledHost, err := hostgroup.gatherHosts(hostname)
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

// Similar to the gather function, this performs a depth first search through
// an inventory's hosts and hostgroups. The variables in the lower/inner hosts and host groups
// will take precedence over the variables in the upper/outer hostgroups
// should a hostgroup name match one of the names passed as a parameter
// all hosts in that hostgroup will be a part of the returned []*Host
func (i Inventory) ExecutionHosts(names []string) ([]*Host, error) {
	hosts := make([]*Host, 0)

	for _, name := range names {
		hostResults, err := i.gatherHost(name)
		if err != nil {
			continue
		}
		hosts = append(hosts, hostResults)
		hostGroupResults, err := i.gatherHostGroups(name)
		hosts = extend(hosts, hostGroupResults)
	}
	if len(hosts) <= 0 {
		return nil, &InventoryError{
			errorCode: HostNotFoundErrorCode,
			Err:       errors.New(fmt.Sprintf("Unable to locate hosts: %v\n", names)),
		}
	}
	return hosts, nil
}

func (i Inventory) gatherHostgroups(name string) ([]*Host, error) {
	hosts := make([]*Host, 0)
	return hosts, nil
}

func (g HostGroup) collect() []*Host {
	hosts := make([]*Host, 0)
	for hostname, host := range g.Hosts {
		newHost := &Host{
			Vars: make(map[string]string, len(host.Vars)),
		}
		for varKey, value := range host.Vars {
			newHost.Vars[varKey] = value
		}
		hosts = append(hosts, newHost)
		
	}
	for _, hostGroup := range g.Children {
		subHostGroupResults := hostGroup.collect()
		hosts = extend(hosts, subHostsGroupResults)
	}
	return hosts
}

func extend(a, b []*Host) []*Host {
	c := make([]*Host, len(a)+len(b))
	index := 0
	for _, value := range a {
		c[index] = value
		index++
	}
	for _, value := range b {
		c[index] = value
		index++
	}
	return c
}
