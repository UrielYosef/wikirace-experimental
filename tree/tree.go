package tree

import (
	"fmt"
	"sort"
)

type Node struct {
	Name     string
	Level    int
	Parent   *Node
	Children *[]*Node
}

func NewTree(name string) *Node {
	tree := &Node{Name: name, Level: 0, Parent: nil, Children: nil}

	return tree
}

func (node *Node) Insert(names []string) {
	node.Children = &[]*Node{}
	for _, name := range names {
		*node.Children = append(*node.Children, &Node{
			Name:     name,
			Level:    node.Level + 1,
			Parent:   node,
			Children: nil})
	}
}

func (node *Node) Depth() int {
	depth := depth(node, 0)

	return depth
}

func (node *Node) PrintRouteToRoot() int {
	fmt.Print("START ")
	depth := printRouteToRoot(node, 0)
	fmt.Println(" END")

	return depth
}

func depth(node *Node, i int) int {
	if node == nil {
		return i - 1
	}
	if node.Children == nil {
		return i
	}

	depths := make([]int, 0)
	for _, child := range *node.Children {
		depths = append(depths, depth(child, i+1))
	}
	if depths == nil || len(depths) == 0 {
		return i
	}

	sortedDepths := sort.IntSlice(depths)
	return sortedDepths[len(sortedDepths)-1]
}

func printRouteToRoot(node *Node, depth int) int {
	if node.Parent == nil {
		fmt.Print(node.Name + " -> ")
		return depth
	} else {
		depth = printRouteToRoot(node.Parent, depth+1)
		fmt.Print(node.Name + " -> ")
	}

	return depth
}
