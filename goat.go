package main

import (
	"flag"
	"fmt"
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

	inventory, err := InventoryFromFilepath(*inventoryFlag)
	if err != nil {
		fmt.Printf("Error when reading inventory file: %v\n", err)
		os.Exit(1)
	}

	playbookPath := flag.Args()[0]
	playbook, err := PlaybookFromFilepath(playbookPath)
	if err != nil {
		fmt.Printf("Error when reading playbook file: %v\n", err)
		os.Exit(1)
	}
	
	results := playbook.Execute(inventory)
	for taskName := range results {
		fmt.Printf("Task: %v\n", taskName)
		for hostname, hostRes := range results[taskName] {
			fmt.Printf("\t%v\n", hostname)
			fmt.Printf("\t\tstdout: %v\n", hostRes.Stdout())
			fmt.Printf("\t\tstderr: %v\n", hostRes.Stderr())
			fmt.Printf("\t\terrors: %v\n", hostRes.Error())			
		}
	}
}
