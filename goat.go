package main

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"flag"
	"os"
)

func main() {

	inventoryFlag := flag.String("inventory", "", "Path to inventory yaml file")
	flag.Parse()
	if inventoryFlag == nil || *inventoryFlag == "" {
		fmt.Printf("Please specify the --inventory flag\n")
		os.Exit(1)
	}

	if flag.NArg() <= 0 {
		fmt.Printf("Usage: goat --inventory <path to inventory>.yaml [playbook yaml]\n")
		os.Exit(1)
	}

	_, err := InventoryFromFilepath(*inventoryFlag)
	if err != nil {
		fmt.Printf("Error when reading inventory file: %v\n", err)
		os.Exit(1)
	}
	
	// An SSH client is represented with a ClientConn.
	//
	// To authenticate with the remote server you must pass at least one
	// implementation of AuthMethod via the Auth field in ClientConfig,
	// and provide a HostKeyCallback.
	config := &ssh.ClientConfig{
		User: "test",
		Auth: []ssh.AuthMethod{
			ssh.Password("test"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "localhost:2222", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/whoami"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())

}
