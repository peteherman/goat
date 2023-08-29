package main

import (
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
	Name    string `yaml:name`
	Command string `yaml:cmd`
}

type TaskResult interface {
	Stdout() string
	Stderr() string
	ReturnCode() int
}

type TaskName string

type PlaybookResult map[TaskName]map[string]TaskResult

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

func (p Playbook) Execute(inventory Inventory) PlaybookResult {

	//executionHosts := inventory.ExecutionHosts(p.Hosts)

	// for _, task := range p.Tasks {

	// }
	return make(PlaybookResult, 0)
}
