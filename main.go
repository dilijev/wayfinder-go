package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Wayfinder struct {
	g     *graph.Mutable
	nodes map[string]int
	rev   map[int]string
	next  int
	state string
}

func NewWayfinder() *Wayfinder {
	return &Wayfinder{
		g:     graph.New(0),
		nodes: make(map[string]int),
		rev:   make(map[int]string),
		next:  0,
		state: "",
	}
}

func (wf *Wayfinder) addNode(description string) {
	if _, exists := wf.nodes[description]; !exists {
		wf.nodes[description] = wf.next
		wf.rev[wf.next] = description
		wf.g.AddVertex()
		wf.next++
	}
}

func (wf *Wayfinder) setCurrentNode(node string) {
	wf.state = node
}

func (wf *Wayfinder) saveState() {
	file, err := os.Create("state.txt")
	if err != nil {
		fmt.Println("Error saving state:", err)
		return
	}
	defer file.Close()
	_, err = file.WriteString(wf.state)
	if err != nil {
		fmt.Println("Error writing state to file:", err)
	}
}

func (wf *Wayfinder) loadState() {
	data, err := os.ReadFile("state.txt")
	if err != nil {
		fmt.Println("Error loading state:", err)
		return
	}
	wf.state = string(data)
}

func (wf *Wayfinder) suggestNodesConcurrent(partial string) []string {
	var suggestions []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	for node := range wf.nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			if strings.Contains(node, partial) {
				mu.Lock()
				suggestions = append(suggestions, node)
				mu.Unlock()
			}
		}(node)
	}

	wg.Wait()
	return suggestions
}

func main() {
	wf := NewWayfinder()
	wf.loadState()

	action := flag.String("action", "", "Action to perform: add-node, add-edge, find-path, suggest, at")
	params := flag.String("params", "", "Parameters for the action")

	flag.Parse()

	switch *action {
	case "at":
		if *params == "" {
			fmt.Println("Please provide a node name.")
			os.Exit(1)
		}
		node := *params
		wf.setCurrentNode(node)
		wf.saveState()
		fmt.Printf("Current node set to '%s'\n", node)

	default:
		fmt.Println("Invalid action. Please use one of the following: add-node, add-edge, find-path, suggest, at")
		os.Exit(1)
	}
}

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

// A function to handle the command line arguments

// Function to handle the requested action no matter whether it's a command line or a REPL
func handleAction(wf *Wayfinder, action string, params string) {
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

	case "to":
		if params == "" {
			fmt.Println("Please provide a node name.")
			os.Exit(1)
		}
		node := params
		path, err := wf.findPath(node)
		if err != nil {
			fmt.Println("Error finding path:", err)
			os.Exit(1)
		}
		fmt.Println("Path:", path)

	default:
		fmt.Println("Invalid action. Please use one of the following: at")
		os.Exit(1)
	}
}

func main() {
	wf := NewWayfinder()
	wf.loadState()

	// If the command line is "repl" then start the REPL
	// Otherwise, parse the command line arguments
	// and execute the specified action.

	// If not enough args, print help
	if len(os.Args) < 2 {
		fmt.Println("Usage: wayfinder [action] [params]")
		os.Exit(1)
	}

	if os.Args[1] == "repl" {
		repl(wf)
	} else {
		wf.handleAction(os.Args[1], strings.Join(os.Args[2:], " "))

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
	}
}
