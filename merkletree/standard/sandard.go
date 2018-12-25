package standard

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math"
	"strings"
	"tree/merkletree"
)

var debug = false

func printf(format string, a ...interface{}) {
	if !debug {
		return
	}
	fmt.Printf(format, a...)
}

type StandardNode struct {
	hash   common.Hash
	index  int
	parent *StandardNode
}

func (node *StandardNode) Hash() string {
	return node.hash.Hex()
}

func (node *StandardNode) Index() int {
	return node.index
}

func (n StandardNode) String() string {
	return n.hash.Hex()
}

// StandardMerkleTree is the most basic implementation of a MerkleTree
type StandardMerkleTree struct {
	nodes [][]*StandardNode
	root  *StandardNode
}

func (tree *StandardMerkleTree) init() {
	tree.nodes = make([][]*StandardNode, 1)
}

func (tree *StandardMerkleTree) resize() {
	leafs := len(tree.nodes[0])
	levels := len(tree.nodes)
	neededLevels := int(math.Ceil(math.Log2(float64(leafs)))) + 1

	if levels < neededLevels {
		n := make([][]*StandardNode, neededLevels)
		copy(n, tree.nodes)
		tree.nodes = n

		printf("StandardMerkleTree resized to %v levels\n", neededLevels)
	}

}

func (tree *StandardMerkleTree) recalculate() (Root *StandardNode) {

	tree.resize()

	levelCount := len(tree.nodes[0])
	level := 0
	tree.nodes[level+1] = make([]*StandardNode, (levelCount/2)+levelCount%2)

	printf("=== N: %v ===\n", levelCount)
	printf("Level: %v, Level count: %v\n", level, levelCount)

	for i := 0; levelCount > 1; i += 2 {
		var left, right *StandardNode

		left = tree.nodes[level][i]

		if i == levelCount-1 { // Odd Nodes level
			right = tree.nodes[level][i]
		} else { // Even Nodes level
			right = tree.nodes[level][i+1]
		}

		node := StandardNode{
			hash:   crypto.Keccak256Hash(left.hash[:], right.hash[:]),
			parent: nil,
			index:  i / 2,
		}

		left.parent = &node
		right.parent = &node

		tree.nodes[level+1][i/2] = &node

		if i+2 >= levelCount {
			levelCount = (levelCount / 2) + levelCount%2
			level++
			if levelCount > 1 {
				tree.nodes[level+1] = make([]*StandardNode, (levelCount/2)+levelCount%2)
				i = -2
			}

			printf("Level: %v, Level count: %v\n", level, levelCount)
		}
	}

	printf("====\n")

	return tree.nodes[level][0]
}

func (tree *StandardMerkleTree) getNodeSibling(level int, index int) *StandardNode {
	nodesCount := len(tree.nodes[level])
	if index%2 == 1 {
		return tree.nodes[level][index-1]
	}

	if index == nodesCount-1 {
		return tree.nodes[level][index]
	}

	return tree.nodes[level][index+1]
}

func (tree *StandardMerkleTree) getLeafSibling(index int) *StandardNode {
	return tree.getNodeSibling(0, index)
}

func (tree *StandardMerkleTree) getIntermediaryHashesByIndex(index int) (intermediaryHashes []*StandardNode) {
	levels := len(tree.nodes)
	if levels < 2 {
		return make([]*StandardNode, 0)
	}
	intermediaryHashes = make([]*StandardNode, 1, levels-1)

	intermediaryHashes[0] = tree.getLeafSibling(index)
	index /= 2

	node := tree.nodes[0][index].parent
	level := 1
	for node.parent != nil {
		intermediaryHashes = append(intermediaryHashes, tree.getNodeSibling(level, index))
		level++
		index /= 2
		node = node.parent
	}

	return intermediaryHashes
}

// Add hashes and inserts data on the next available slot in the tree.
// Also recalculates and recalibrates the tree.
// Returns the index it was inserted and the hash of the new data
func (tree *StandardMerkleTree) Add(data []byte) (index int, hash string) {
	index = len(tree.nodes[0])

	leaf := &(StandardNode{
		crypto.Keccak256Hash(data),
		index,
		nil,
	})

	tree.nodes[0] = append(tree.nodes[0], leaf)

	if index == 0 {
		tree.root = leaf
	} else {
		tree.root = tree.recalculate()
	}
	return index, leaf.Hash()
}

// IntermediaryHashesByIndex returns all hashes needed to produce the root from the comming index
func (tree *StandardMerkleTree) IntermediaryHashesByIndex(index int) (intermediaryHashes []string) {
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
func (tree *StandardMerkleTree) ValidateExistence(original []byte, index int, intermediaryHashes []string) bool {
	leafHash := crypto.Keccak256Hash(original)

	treeLeaf := tree.nodes[0][index]

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

	return tempBHash.Big().Cmp(tree.root.hash.Big()) == 0

}

// Root returns the hash of the root of the tree
func (tree *StandardMerkleTree) Root() string {
	return tree.root.Hash()
}

// Length returns the count of the tree leafs
func (tree *StandardMerkleTree) Length() int {
	return len(tree.nodes[0])
}

// String returns human readable version of the tree
func (tree StandardMerkleTree) String() string {
	b := strings.Builder{}

	l := len(tree.nodes)

	for i := l - 1; i >= 0; i-- {
		ll := len(tree.nodes[i])
		b.WriteString(fmt.Sprintf("Level: %v, Count: %v\n", i, ll))
		for k := 0; k < ll; k++ {
			b.WriteString(fmt.Sprintf("%v\t", tree.nodes[i][k].Hash()))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// HashAt returns the hash at given index
func (tree *StandardMerkleTree) HashAt(index int) string {
	return tree.nodes[0][index].Hash()
}

// MarshalJSON Creates JSON version of the needed fields of the tree
func (tree *StandardMerkleTree) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("{\"root\":\"%v\", \"length\":%v}", tree.Root(), tree.Length())
	return []byte(res), nil
}

// NewTree returns a pointer to an initialized StandardMerkleTree
func NewTree() merkletree.MerkleTree {
	var tree StandardMerkleTree
	tree.init()

	return &tree
}
