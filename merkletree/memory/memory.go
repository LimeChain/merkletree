// Package memory implements merkle tree stored in the memory of the system
package memory

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math"
	"strings"
	"sync"
)

const (
	outOfBounds = "Incorrect index - Index out of bounds"
)

// Node is implementation of types.Node and representation of a single node or leaf in the merkle tree
type Node struct {
	hash   common.Hash
	index  int
	Parent *Node
}

// Hash returns the string representation of the hash of the node
func (node *Node) Hash() string {
	return node.hash.Hex()
}

// Index returns the index of this node in its level
func (node *Node) Index() int {
	return node.index
}

// String returns the hash of this node. Alias to Hash()
func (node Node) String() string {
	return node.Hash()
}

// MerkleTree is the most basic implementation of a MerkleTree
type MerkleTree struct {
	Nodes    [][]*Node
	RootNode *Node
	mutex    sync.RWMutex
}

func (tree *MerkleTree) init() {
	tree.Nodes = make([][]*Node, 1)
}

func (tree *MerkleTree) resizeVertically() {
	leafs := len(tree.Nodes[0])
	levels := len(tree.Nodes)
	neededLevels := int(math.Ceil(math.Log2(float64(leafs)))) + 1

	if levels < neededLevels {
		n := make([][]*Node, neededLevels)
		copy(n, tree.Nodes)
		tree.Nodes = n
	}

}

func (tree *MerkleTree) propagateChange() (root *Node) {

	tree.resizeVertically()

	levels := len(tree.Nodes)

	lastNodeSibling := func(nodes []*Node, length int) *Node {
		if length%2 == 0 {
			// The added node completed a pair - take the other half
			return nodes[length-2]
		}
		// The added node created new pair - duplicate itself
		return nodes[length-1]
	}

	createParent := func(left, right *Node) *Node {
		parentNode := &Node{
			hash:   crypto.Keccak256Hash(left.hash[:], right.hash[:]),
			Parent: nil,
			index:  right.index / 2, // Parent index is always the current node index divided by two
		}

		left.Parent = parentNode
		right.Parent = parentNode

		return parentNode
	}

	updateParentLevel := func(parent *Node, parentLevel []*Node) []*Node {
		nextLevelLen := len(parentLevel)
		if parent.index == nextLevelLen { // If the leafs are now odd, The parent needs to expand the level
			parentLevel = append(parentLevel, parent)
		} else {
			parentLevel[parent.index] = parent // If the leafs are now even, The parent is just replaced
		}

		return parentLevel
	}

	for i := 0; i < (levels - 1); i++ {
		var left, right *Node

		levelLen := len(tree.Nodes[i])

		right = tree.Nodes[i][levelLen-1]               // Last inserted node
		left = lastNodeSibling(tree.Nodes[i], levelLen) // Either the other half or himself

		parentNode := createParent(left, right) // Create parent hashing the two

		tree.Nodes[i+1] = updateParentLevel(parentNode, tree.Nodes[i+1]) // Update the parent level

	}

	root = tree.Nodes[levels-1][0]

	return root
}

func (tree *MerkleTree) getNodeSibling(level int, index int) *Node {
	nodesCount := len(tree.Nodes[level])
	if index%2 == 1 {
		return tree.Nodes[level][index-1]
	}

	if index == nodesCount-1 {
		return tree.Nodes[level][index]
	}

	return tree.Nodes[level][index+1]
}

func (tree *MerkleTree) getLeafSibling(index int) *Node {
	return tree.getNodeSibling(0, index)
}

func (tree *MerkleTree) getIntermediaryHashesByIndex(index int) (intermediaryHashes []*Node) {
	levels := len(tree.Nodes)
	if levels < 2 {
		return make([]*Node, 0)
	}
	intermediaryHashes = make([]*Node, 1, levels-1)

	intermediaryHashes[0] = tree.getLeafSibling(index)
	index /= 2

	node := tree.Nodes[0][index].Parent
	level := 1
	for node.Parent != nil {
		intermediaryHashes = append(intermediaryHashes, tree.getNodeSibling(level, index))
		level++
		index /= 2
		node = node.Parent
	}

	return intermediaryHashes
}

// Add hashes and inserts data on the next available slot in the tree.
// Also recalculates and recalibrates the tree.
// Returns the index it was inserted and the hash of the new data
func (tree *MerkleTree) Add(data []byte) (index int, hash string) {
	tree.mutex.RLock()
	index = len(tree.Nodes[0])

	leaf := &Node{
		crypto.Keccak256Hash(data),
		index,
		nil,
	}

	tree.Nodes[0] = append(tree.Nodes[0], leaf)

	if index == 0 {
		tree.RootNode = leaf
	} else {
		tree.RootNode = tree.propagateChange()
	}
	tree.mutex.RUnlock()
	return index, leaf.Hash()
}

// IntermediaryHashesByIndex returns all hashes needed to produce the root from the comming index
func (tree *MerkleTree) IntermediaryHashesByIndex(index int) (intermediaryHashes []string, err error) {
	if index >= len(tree.Nodes[0]) {
		return nil, errors.New(outOfBounds)
	}
	hashes := tree.getIntermediaryHashesByIndex(index)
	intermediaryHashes = make([]string, len(hashes))
	for i, h := range hashes {
		intermediaryHashes[i] = h.Hash()
	}

	return intermediaryHashes, nil
}

// ValidateExistence emulates how third party would validate the data
// Given original data, the index it is supposed to be and the intermediaryHashes to the root
// Validates that this is the correct data for that slot
// In production you can just check the HashAt and hash the original data yourself
func (tree *MerkleTree) ValidateExistence(original []byte, index int, intermediaryHashes []string) (result bool, err error) {
	if index >= len(tree.Nodes[0]) {
		return false, errors.New(outOfBounds)
	}
	leafHash := crypto.Keccak256Hash(original)

	treeLeaf := tree.Nodes[0][index]

	if leafHash.Big().Cmp(treeLeaf.hash.Big()) != 0 {
		return false, nil
	}

	tempBHash := leafHash

	for _, h := range intermediaryHashes {
		oppositeHash := common.HexToHash(h)

		if index%2 == 0 {
			tempBHash = crypto.Keccak256Hash(tempBHash[:], oppositeHash[:])
		} else {
			tempBHash = crypto.Keccak256Hash(oppositeHash[:], tempBHash[:])
		}

		index /= 2
	}

	return tempBHash.Big().Cmp(tree.RootNode.hash.Big()) == 0, nil

}

// Root returns the hash of the root of the tree
func (tree *MerkleTree) Root() string {
	return tree.RootNode.Hash()
}

// Length returns the count of the tree leafs
func (tree *MerkleTree) Length() int {
	return len(tree.Nodes[0])
}

// String returns human readable version of the tree
func (tree *MerkleTree) String() string {
	b := strings.Builder{}

	l := len(tree.Nodes)

	for i := l - 1; i >= 0; i-- {
		ll := len(tree.Nodes[i])
		b.WriteString(fmt.Sprintf("Level: %v, Count: %v\n", i, ll))
		for k := 0; k < ll; k++ {
			b.WriteString(fmt.Sprintf("%v\t", tree.Nodes[i][k].Hash()))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// HashAt returns the hash at given index
func (tree *MerkleTree) HashAt(index int) (string, error) {
	if index >= len(tree.Nodes[0]) {
		return "", errors.New(outOfBounds)
	}
	return tree.Nodes[0][index].Hash(), nil
}

// MarshalJSON Creates JSON version of the needed fields of the tree
func (tree *MerkleTree) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("{\"root\":\"%v\", \"length\":%v}", tree.Root(), tree.Length())
	return []byte(res), nil
}

// NewMerkleTree returns a pointer to an initialized MerkleTree
func NewMerkleTree() *MerkleTree {
	var tree MerkleTree
	tree.init()

	return &tree
}
