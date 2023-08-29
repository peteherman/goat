package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Playbook struct {
	Name  string            `yaml:name`
	Hosts []string          `yaml:hosts`
	Vars  map[string]string `yaml:vars,omit=empty`
	Tasks []Task            `yaml:tasks`
}

type Task map[string]any

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
		Tasks: make([]Task, 0),
	}
	err := yaml.Unmarshal(contents, &playbook)
	if err != nil {
		return Playbook{}, err
	}
	return playbook, nil
}
