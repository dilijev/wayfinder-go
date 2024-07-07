package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	// yourbasic.org/graph as graph
	"github.com/yourbasic/graph"
)

type Wayfinder struct {
	g     *graph.Mutable
	nodes map[string]int
	rev   map[int]string
	next  int
	state string
}

type serializableWayfinder struct {
	Nodes map[string]int `json:"nodes"`
	Rev   map[int]string `json:"rev"`
	Next  int            `json:"next"`
	State string         `json:"state"`
	Edges [][2]int       `json:"edges"` // Represents graph edges
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

func (wf *Wayfinder) MarshalJSON() ([]byte, error) {
	// Create a list to store the edges
	var edges [][2]int
	for v := 0; v < wf.g.Order(); v++ {
		wf.g.Visit(v, func(w int, c int64) bool {
			edges = append(edges, [2]int{v, w})
			return true // continue visiting neighbors
		})
	}

	// Create a serializable version of Wayfinder
	sw := serializableWayfinder{
		Nodes: wf.nodes,
		Rev:   wf.rev,
		Next:  wf.next,
		State: wf.state,
		Edges: edges,
	}

	return json.Marshal(sw)
}

func (wf *Wayfinder) UnmarshalJSON(data []byte) error {
	// Unmarshal into the serializable struct
	var sw serializableWayfinder
	if err := json.Unmarshal(data, &sw); err != nil {
		return err
	}

	// Recreate the graph
	wf.g = graph.New(len(sw.Nodes))
	for _, edge := range sw.Edges {
		wf.g.AddCost(edge[0], edge[1], 1)
	}

	// Restore other fields
	wf.nodes = sw.Nodes
	wf.rev = sw.Rev
	wf.next = sw.Next
	wf.state = sw.State

	return nil
}

func (wf *Wayfinder) addNode(description string) {
	if _, exists := wf.nodes[description]; !exists {
		wf.nodes[description] = wf.next
		wf.rev[wf.next] = description
		// wf.g.AddVertex()
		wf.next++
	}
}

func (wf *Wayfinder) setCurrentNode(node string) {
	wf.state = node
}

// getCurrentNode
func (wf *Wayfinder) getCurrentNode() string {
	return wf.state
}

func (wf *Wayfinder) saveState() {
	file, err := os.Create("state.json")
	if err != nil {
		fmt.Println("Error saving state:", err)
		return
	}
	defer file.Close()

	// define bytes
	var bytes []byte
	bytes, err = wf.MarshalJSON()
	if err != nil {
		fmt.Println("Error marshaling state:", err)
		return
	}

	_, err = file.Write(bytes)
	if err != nil {
		fmt.Println("Error writing state to file:", err)
		return
	}
}

func (wf *Wayfinder) loadState() {
	data, err := os.ReadFile("state.json")
	if err != nil {
		fmt.Println("Error loading state:", err)
		return
	}

	err = wf.UnmarshalJSON(data)
	if err != nil {
		fmt.Println("Error unmarshaling state:", err)
	}
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

func (wf *Wayfinder) findPath(start, end string) ([]string, error) {
	id1, ok1 := wf.nodes[start]
	id2, ok2 := wf.nodes[end]
	if !ok1 || !ok2 {
		return nil, errors.New("one or both nodes not found")
	}

	// Find the shortest path using Dijkstra's algorithm
	path, _ := graph.ShortestPath(wf.g, id1, id2)
	if len(path) == 0 {
		return nil, errors.New("no path found")
	}

	var result []string
	// what does range do?
	// range returns the key of the map
	// for each key in the map, append the value to the result
	// result is a slice of strings
	// the slice is the path from start to end
	for id := range path {
		// convert id from int64 to int
		id := int(id)
		result = append(result, wf.rev[id])
	}
	return result, nil
}

// A function to handle the command line arguments

// Function to handle the requested action no matter whether it's a command line or a REPL
func (wf *Wayfinder) handleAction(action string, params string) {
	switch action {
	case "at":
		if params == "" {
			fmt.Println("Please provide a node name.")
			os.Exit(1)
		}
		at_node := params
		wf.setCurrentNode(at_node)
		wf.saveState()
		fmt.Printf("Current node set to '%s'\n", at_node)

	case "to":
		if params == "" {
			fmt.Println("Please provide a node name.")
			os.Exit(1)
		}
		to_node := params
		path, err := wf.findPath(wf.getCurrentNode(), to_node)
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
		case "suggest":
			suggestions := wf.suggestNodesConcurrent(params)
			fmt.Println("Suggestions:", strings.Join(suggestions, ", "))

		case "exit":
			fmt.Println("Exiting REPL.")
			return

		default:
			wf.handleAction(command, params)
		}
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
