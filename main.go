package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
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
		g:     graph.New(200),
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
	nodes := len(sw.Nodes)
	if nodes < 200 {
		nodes = 200
	}
	wf.g = graph.New(nodes)

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
	fmt.Println("Saving state to file:", g_ms.Savefile)

	file, err := os.Create(g_ms.Savefile)
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

// create default state if loadState fails
func CreateBlankWayfinderStateFile() {
	wf := NewWayfinder()
	wf.saveState()
}

func (wf *Wayfinder) loadState() error {
	data, err := os.ReadFile(g_ms.Savefile)
	if err != nil {
		fmt.Println("Error loading state:", err)
		fmt.Println("Creating blank state file.")
		CreateBlankWayfinderStateFile()
		return nil
	}

	err = wf.UnmarshalJSON(data)
	if err != nil {
		fmt.Println("Error unmarshaling state:", err)
		return err
	}

	return nil
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

func (wf *Wayfinder) addEdge(from, to string, weight int) {
	wf.addNode(from)
	wf.addNode(to)

	id1 := wf.nodes[from]
	id2 := wf.nodes[to]
	wf.g.AddCost(id1, id2, int64(weight))

	fmt.Printf("Added edge from '%s' to '%s' with weight %d\n", from, to, weight)

	// Save the state after adding an edge
	wf.saveState()
}

func (wf *Wayfinder) findPath(start, end string) ([]string, error) {
	id1, ok1 := wf.nodes[start]
	id2, ok2 := wf.nodes[end]
	if !ok1 || !ok2 {
		return nil, errors.New("one or both nodes not found")
	}

	fmt.Println("start:", start, "end:", end)

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
	for x := range path {
		// convert id from int64 to int
		id := path[x]

		label := wf.rev[id]
		fmt.Println("id:", id, "rev[id]:", label)

		result = append(result, label)
	}
	return result, nil
}

// A function to handle the command line arguments

// Function to handle the requested action no matter whether it's a command line or a REPL
func (wf *Wayfinder) handleAction(action string, params []string) {
	switch action {
	case "at":
		if len(params) == 0 {
			fmt.Println("Please provide a node name.")
			return
		}
		at_node := params[0]
		wf.setCurrentNode(at_node)
		wf.saveState()
		fmt.Printf("Current node set to '%s'\n", at_node)

	case "to":
		if len(params) == 0 {
			fmt.Println("Please provide a node name.")
			return
		}
		to_node := params[0]
		path, err := wf.findPath(wf.getCurrentNode(), to_node)
		if err != nil {
			fmt.Println("Error finding path:", err)
			return
		}
		fmt.Println("Path:", path)

	case "add":
		if len(params) < 2 {
			fmt.Println("Please provide from_node and to_node.")
			return
		}
		// the first param
		from_node := params[0]
		to_node := params[1]
		// get params[2] if it exists, otherwise default to "1"
		weight := 1
		if len(params) > 2 {
			var err error
			weight, err = strconv.Atoi(params[2])
			if err != nil {
				fmt.Println("Invalid weight:", err)
				return
			}
		}

		wf.addEdge(from_node, to_node, weight)

	case "nodes":
		fmt.Println("Nodes:")

		// add all the nodes to a list
		var nodes []string
		for node := range wf.nodes {
			// append the node to the list
			nodes = append(nodes, node)
		}

		// sort the list
		sort.Strings(nodes)
		for _, node := range nodes {
			fmt.Println(node)
		}

	case "edges":
		// list edges
		fmt.Println("Edges:")
		for v := 0; v < wf.g.Order(); v++ {
			wf.g.Visit(v, func(w int, c int64) bool {
				fmt.Printf("%s [%d] -> %s [%d] (%d)\n", wf.rev[v], v, wf.rev[w], w, c)
				return true // continue visiting neighbors
			})
		}

	default:
		fmt.Println("Invalid action. Please use one of the following: at, to, add, nodes, edges, exit")
		return
	}
}

func repl(wf *Wayfinder) {
	fmt.Println("Starting REPL. Type 'exit' to quit.")
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
		params := parts[1:]

		switch command {
		// case "suggest":
		// 	suggestions := wf.suggestNodesConcurrent(params)
		// 	fmt.Println("Suggestions:", strings.Join(suggestions, ", "))

		case "exit":
			wf.saveState()
			fmt.Println("Exiting REPL.")
			return

		default:
			wf.handleAction(command, params)
		}
	}
}

// JSON serializable
type MetaState struct {
	// the current save file
	Savefile string `json:"savefile"`
}

func (ms *MetaState) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms)
}

func (ms *MetaState) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, ms)
}

func (ms *MetaState) saveState() {
	file, err := os.Create("metastate.json")
	if err != nil {
		fmt.Println("Error saving metastate:", err)
		return
	}
	defer file.Close()

	// define bytes
	var bytes []byte
	bytes, err = ms.MarshalJSON()
	if err != nil {
		fmt.Println("Error marshaling metastate:", err)
		return
	}

	_, err = file.Write(bytes)
	if err != nil {
		fmt.Println("Error writing metastate to file:", err)
		return
	}
}

func ForceOverwriteMetaState() *MetaState {
	ms := MetaState{
		Savefile: "state.json",
	}
	ms.saveState()
	return &ms
}

func LoadMetaState(filename string) *MetaState {
	// if filename does not exist, or the contents do not parse as JSON, write a new one
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("No metastate file found. Creating metadata.json")
		return ForceOverwriteMetaState()
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error loading metastate.json. Resetting to default.", err)
		return ForceOverwriteMetaState()
	}

	var ms MetaState
	err = ms.UnmarshalJSON(data)
	if err != nil {
		fmt.Println("Error unmarshaling metastate. Resetting to default.", err)
		return ForceOverwriteMetaState()
	}

	return &ms
}

// global g_ms *MetaState
var g_ms *MetaState

func main() {
	// TODO get this dynamic loading to work instead of forced "state.json"
	// g_ms = LoadMetaState("metastate.json")
	// if g_ms == nil {
	// 	fmt.Println("Error loading metastate.")
	// 	return
	// }

	g_ms = &MetaState{Savefile: "state.json"}

	wf := NewWayfinder()
	err := wf.loadState()
	if err != nil {
		fmt.Println("Error loading state:", err)
		return
	}

	// If the command line is "repl" then start the REPL
	// Otherwise, parse the command line arguments
	// and execute the specified action.

	// If not enough args, print help
	if len(os.Args) < 2 {
		fmt.Println("Usage: wayfinder [action] [params]")
		os.Exit(1)
	}

	// consider a top level context file separate from save file
	// this file tells us the path to the save file to use
	// if not set, start with the default "state.json" in the current directory

	if os.Args[1] == "savefile" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: wayfinder savefile [filename]")
			os.Exit(1)
		}

		g_ms.Savefile = os.Args[2]
		g_ms.saveState()
	} else if os.Args[1] == "repl" {
		repl(wf)
	} else {
		action := os.Args[1]
		params := os.Args[2:]
		wf.handleAction(action, params)
	}
}
