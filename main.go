/*
I'd like to make a shortest path finder with a CLI to dynamically add links and query. Each node will be a text description. I don't necessarily want to have a set of texts to choose from but I'd be interested to have that as an option such that the CLI will suggest options from the pre-populated list for the item you tried to enter. The main inspiration is wayfinding in a decoupled entrance randomizer of Ocarina of Time and Majora's Mask
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "wayfinder"
	app.Usage = "find the shortest path between two nodes"
	app.Action = func(c *cli.Context) error {
		fmt.Println("Hello friend!")
		return nil
	}
	app.Run(os.Args)
}

func addLink() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the first node: ")
	node1, _ := reader.ReadString('\n')
	fmt.Print("Enter the second node: ")
	node2, _ := reader.ReadString('\n')
	fmt.Print("Enter the distance between the nodes: ")
	distance, _ := reader.ReadString('\n')
}

func query() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the starting node: ")
	start, _ := reader.ReadString('\n')
	fmt.Print("Enter the ending node: ")
	end, _ := reader.ReadString('\n')
}

func suggest() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the first few characters of the node: ")
	node, _ := reader.ReadString('\n')
}

func shortestPath() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the starting node: ")
	start, _ := reader.ReadString('\n')
	fmt.Print("Enter the ending node: ")
	end, _ := reader.ReadString('\n')
}
