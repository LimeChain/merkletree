package memory

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math"
	"strings"
	"sync"
)

var debug = false

func printf(format string, a ...interface{}) {
	if !debug {
		return
	}
	fmt.Printf(format, a...)
}

type MemoryNode struct {
	hash   common.Hash
	index  int
	Parent *MemoryNode
}

func (node *MemoryNode) Hash() string {
	return node.hash.Hex()
}

func (node *MemoryNode) Index() int {
	return node.index
}

func (n MemoryNode) String() string {
	return n.hash.Hex()
}

// MemoryMerkleTree is the most basic implementation of a MemoryMerkleTree
type MemoryMerkleTree struct {
	Nodes [][]*MemoryNode
	_Root *MemoryNode
	mutex sync.RWMutex
}

func (tree *MemoryMerkleTree) init() {
	tree.Nodes = make([][]*MemoryNode, 1)
}

func (tree *MemoryMerkleTree) resizeVertically() {
	leafs := len(tree.Nodes[0])
	levels := len(tree.Nodes)
	neededLevels := int(math.Ceil(math.Log2(float64(leafs)))) + 1

	if levels < neededLevels {
		n := make([][]*MemoryNode, neededLevels)
		copy(n, tree.Nodes)
		tree.Nodes = n

		printf("MemoryMerkleTree resized to %v levels\n", neededLevels)
	}

}

func (tree *MemoryMerkleTree) propagateChange() (root *MemoryNode) {

	tree.resizeVertically()

	levels := len(tree.Nodes)

	printf("Levels %v\n", levels)

	lastNodeSibling := func(nodes []*MemoryNode, length int) *MemoryNode {
		if length%2 == 0 {
			// The added node completed a pair - take the other half
			return nodes[length-2]
		}
		// The added node created new pair - duplicate itself
		return nodes[length-1]
	}

	createParent := func(left, right *MemoryNode) *MemoryNode {
		parentNode := &MemoryNode{
			hash:   crypto.Keccak256Hash(left.hash[:], right.hash[:]),
			Parent: nil,
			index:  right.index / 2, // Parent index is always the current node index divided by two
		}

		left.Parent = parentNode
		right.Parent = parentNode

		return parentNode
	}

	updateParentLevel := func(parent *MemoryNode, parentLevel []*MemoryNode) {
		nextLevelLen := len(parentLevel)
		if parent.index == nextLevelLen { // If the leafs are now odd, The parent needs to expand the level
			parentLevel = append(parentLevel, parent)
		} else {
			parentLevel[parent.index] = parent // If the leafs are now even, The parent is just replaced
		}
	}

	for i := 0; i < (levels - 1); i++ {
		var left, right *MemoryNode

		levelLen := len(tree.Nodes[i])

		right = tree.Nodes[i][levelLen-1]               // Last inserted node
		left = lastNodeSibling(tree.Nodes[i], levelLen) // Either the other half or himself

		parentNode := createParent(left, right) // Create parent hashing the two

		updateParentLevel(parentNode, tree.Nodes[i+1]) // Update the parent level

		printf("Level: %v, Level count: %v\n", i, levelLen)
	}

	printf("====\n")

	root = tree.Nodes[levels-1][0]

	return root
}

func (tree *MemoryMerkleTree) getNodeSibling(level int, index int) *MemoryNode {
	nodesCount := len(tree.Nodes[level])
	if index%2 == 1 {
		return tree.Nodes[level][index-1]
	}

	if index == nodesCount-1 {
		return tree.Nodes[level][index]
	}

	return tree.Nodes[level][index+1]
}

func (tree *MemoryMerkleTree) getLeafSibling(index int) *MemoryNode {
	return tree.getNodeSibling(0, index)
}

func (tree *MemoryMerkleTree) getIntermediaryHashesByIndex(index int) (intermediaryHashes []*MemoryNode) {
	levels := len(tree.Nodes)
	if levels < 2 {
		return make([]*MemoryNode, 0)
	}
	intermediaryHashes = make([]*MemoryNode, 1, levels-1)

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
func (tree *MemoryMerkleTree) Add(data []byte) (index int, hash string) {
	tree.mutex.RLock()
	index = len(tree.Nodes[0])

	leaf := &MemoryNode{
		crypto.Keccak256Hash(data),
		index,
		nil,
	}

	tree.Nodes[0] = append(tree.Nodes[0], leaf)

	if index == 0 {
		tree._Root = leaf
	} else {
		tree._Root = tree.propagateChange()
	}
	tree.mutex.RUnlock()
	return index, leaf.Hash()
}

// IntermediaryHashesByIndex returns all hashes needed to produce the root from the comming index
func (tree *MemoryMerkleTree) IntermediaryHashesByIndex(index int) (intermediaryHashes []string) {
	hashes := tree.getIntermediaryHashesByIndex(index)
	intermediaryHashes = make([]string, len(hashes))
	for i, h := range hashes {
		intermediaryHashes[i] = h.Hash()
	}

	return intermediaryHashes
}

// ValidateExistence emulates how third party would validate the data
// Given original data, the index it is supposed to be and the intermediaryHashes to the root
// Validates that this is the correct data for that slot
// In production you can just check the HashAt and hash the original data yourself
func (tree *MemoryMerkleTree) ValidateExistence(original []byte, index int, intermediaryHashes []string) bool {
	leafHash := crypto.Keccak256Hash(original)

	treeLeaf := tree.Nodes[0][index]

	if leafHash.Big().Cmp(treeLeaf.hash.Big()) != 0 {
		return false
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

	return tempBHash.Big().Cmp(tree._Root.hash.Big()) == 0

}

// Root returns the hash of the root of the tree
func (tree *MemoryMerkleTree) Root() string {
	return tree._Root.Hash()
}

// Length returns the count of the tree leafs
func (tree *MemoryMerkleTree) Length() int {
	return len(tree.Nodes[0])
}

// String returns human readable version of the tree
func (tree MemoryMerkleTree) String() string {
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
func (tree *MemoryMerkleTree) HashAt(index int) string {
	return tree.Nodes[0][index].Hash()
}

// MarshalJSON Creates JSON version of the needed fields of the tree
func (tree *MemoryMerkleTree) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("{\"root\":\"%v\", \"length\":%v}", tree.Root(), tree.Length())
	return []byte(res), nil
}

// NewTree returns a pointer to an initialized MemoryMerkleTree
func NewTree() *MemoryMerkleTree {
	var tree MemoryMerkleTree
	tree.init()

	return &tree
}
