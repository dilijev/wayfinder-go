package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	// "sync"
)

func repl(wf *Wayfinder) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("way> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		parts := strings.Split(input, " ")
		command := parts[0]
		params := strings.Join(parts[1:], " ")

		switch command {
		case "at":
			if params == "" {
				fmt.Println("Please provide a node name.")
				continue
			}
			node := params
			wf.setCurrentNode(node)
			wf.saveState()
			fmt.Printf("Current node set to '%s'\n", node)

		case "suggest":
			suggestions := wf.suggestNodesConcurrent(params)
			fmt.Println("Suggestions:", strings.Join(suggestions, ", "))

		case "exit":
			fmt.Println("Exiting REPL.")
			return

		default:
			fmt.Println("Unknown command:", command)
		}
	}
}

func main() {
	wf := NewWayfinder()
	wf.loadState()

	if len(os.Args) > 1 {
		action := os.Args[1]
		params := strings.Join(os.Args[2:], " ")

		switch action {
		case "at":
			if params == "" {
				fmt.Println("Please provide a node name.")
				os.Exit(1)
			}
			node := params
			wf.setCurrentNode(node)
			wf.saveState()
			fmt.Printf("Current node set to '%s'\n", node)

		default:
			fmt.Println("Invalid action. Please use one of the following: at")
			os.Exit(1)
		}
	} else {
		repl(wf)
	}
}
