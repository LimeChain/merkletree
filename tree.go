package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math"
	"strings"
)

var debug = false

type Node struct {
	index  int
	hash   common.Hash
	parent *Node
}

type Tree struct {
	nodes [][]*Node
	root  *Node
}

func printf(format string, a ...interface{}) {
	if !debug {
		return
	}
	fmt.Printf(format, a...)
}

func (tree *Tree) init() {
	tree.nodes = make([][]*Node, 1)
}

func (tree *Tree) insert(data []byte) (index int, leaf *Node, root *Node) {
	index = len(tree.nodes[0])

	leaf = &(Node{
		index,
		crypto.Keccak256Hash(data),
		nil,
	})

	tree.nodes[0] = append(tree.nodes[0], leaf)

	if index == 0 {
		tree.root = leaf
		return index, leaf, tree.root
	}

	root = tree.recalculate()
	tree.root = root

	return index, leaf, leaf
}

func (tree *Tree) resize() {
	leafs := len(tree.nodes[0])
	levels := len(tree.nodes)
	neededLevels := int(math.Ceil(math.Log2(float64(leafs)))) + 1

	if levels < neededLevels {
		n := make([][]*Node, neededLevels)
		copy(n, tree.nodes)
		tree.nodes = n

		printf("Tree resized to %v levels\n", neededLevels)
	}

}

func (tree *Tree) recalculate() (root *Node) {

	tree.resize()

	levelCount := len(tree.nodes[0])
	level := 0
	tree.nodes[level+1] = make([]*Node, (levelCount/2)+levelCount%2)

	printf("=== N: %v ===\n", levelCount)
	printf("Level: %v, Level count: %v\n", level, levelCount)

	for i := 0; levelCount > 1; i += 2 {
		var left, right *Node

		left = tree.nodes[level][i]

		if i == levelCount-1 { // Odd nodes level
			right = tree.nodes[level][i]
		} else { // Even nodes level
			right = tree.nodes[level][i+1]
		}

		node := Node{
			hash:   crypto.Keccak256Hash(left.hash[:], right.hash[:]),
			parent: nil,
		}

		left.parent = &node
		right.parent = &node

		tree.nodes[level+1][i/2] = &node

		if i+2 >= levelCount {
			levelCount = (levelCount / 2) + levelCount%2
			level++
			if levelCount > 1 {
				tree.nodes[level+1] = make([]*Node, (levelCount/2)+levelCount%2)
				i = -2
			}

			printf("Level: %v, Level count: %v\n", level, levelCount)
		}
	}

	printf("====\n")

	return tree.nodes[level][0]
}

func (tree Tree) String() string {
	b := strings.Builder{}

	l := len(tree.nodes)

	for i := l - 1; i >= 0; i-- {
		ll := len(tree.nodes[i])
		b.WriteString(fmt.Sprintf("Level: %v, Count: %v\n", i, ll))
		for k := 0; k < ll; k++ {
			b.WriteString(fmt.Sprintf("%v\t", tree.nodes[i][k].hash.Hex()))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func initTree(elements int) Tree {
	var tree Tree
	tree.init()

	for i := 0; i < elements; i++ {
		tree.insert([]byte("hello" + string(i)))
	}
	fmt.Printf("%v\n", tree)

	return tree
}

func main() {
	tree := initTree(25)
	fmt.Println(tree.root.hash.Hex())
}
