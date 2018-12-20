package merkletree

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math"
	"strings"
)

var debug = false

func printf(format string, a ...interface{}) {
	if !debug {
		return
	}
	fmt.Printf(format, a...)
}

func (node *Node) toHash() string {
	return node.Hash.Hex()
}

type Node struct {
	Index  int
	Hash   common.Hash
	Parent *Node
}

type MerkleTree struct {
	Leafs []*Node
	Nodes [][]*Node
	Root  *Node
}

func (tree *MerkleTree) init() {
	tree.Nodes = make([][]*Node, 1)
}

func (tree *MerkleTree) resize() {
	leafs := len(tree.Nodes[0])
	levels := len(tree.Nodes)
	neededLevels := int(math.Ceil(math.Log2(float64(leafs)))) + 1

	if levels < neededLevels {
		n := make([][]*Node, neededLevels)
		copy(n, tree.Nodes)
		tree.Nodes = n

		printf("MerkleTree resized to %v levels\n", neededLevels)
	}

}

func (tree *MerkleTree) recalculate() (Root *Node) {

	tree.resize()

	levelCount := len(tree.Nodes[0])
	level := 0
	tree.Nodes[level+1] = make([]*Node, (levelCount/2)+levelCount%2)

	printf("=== N: %v ===\n", levelCount)
	printf("Level: %v, Level count: %v\n", level, levelCount)

	for i := 0; levelCount > 1; i += 2 {
		var left, right *Node

		left = tree.Nodes[level][i]

		if i == levelCount-1 { // Odd Nodes level
			right = tree.Nodes[level][i]
		} else { // Even Nodes level
			right = tree.Nodes[level][i+1]
		}

		node := Node{
			Hash:   crypto.Keccak256Hash(left.Hash[:], right.Hash[:]),
			Parent: nil,
		}

		left.Parent = &node
		right.Parent = &node

		tree.Nodes[level+1][i/2] = &node

		if i+2 >= levelCount {
			levelCount = (levelCount / 2) + levelCount%2
			level++
			if levelCount > 1 {
				tree.Nodes[level+1] = make([]*Node, (levelCount/2)+levelCount%2)
				i = -2
			}

			printf("Level: %v, Level count: %v\n", level, levelCount)
		}
	}

	printf("====\n")

	return tree.Nodes[level][0]
}

func (tree *MerkleTree) getNodeSibling(level int, index int) *Node {
	node := len(tree.Nodes[level])
	if index >= node {
		// TODO throw error
	}
	if index%2 == 1 {
		return tree.Nodes[level][index-1]
	}

	if index == node-1 {
		return tree.Nodes[level][index]
	}

	return tree.Nodes[level][index+1]
}

func (tree *MerkleTree) getLeafSibling(index int) *Node {
	return tree.getNodeSibling(0, index)
}

func (tree MerkleTree) String() string {
	b := strings.Builder{}

	l := len(tree.Nodes)

	for i := l - 1; i >= 0; i-- {
		ll := len(tree.Nodes[i])
		b.WriteString(fmt.Sprintf("Level: %v, Count: %v\n", i, ll))
		for k := 0; k < ll; k++ {
			b.WriteString(fmt.Sprintf("%v\t", tree.Nodes[i][k].Hash.Hex()))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (tree *MerkleTree) Add(data []byte) (index int, leaf *Node, root *Node) {
	index = len(tree.Nodes[0])

	leaf = &(Node{
		index,
		crypto.Keccak256Hash(data),
		nil,
	})

	tree.Nodes[0] = append(tree.Nodes[0], leaf)

	if index == 0 {
		tree.Root = leaf
		return index, leaf, tree.Root
	}

	root = tree.recalculate()
	tree.Root = root
	tree.Leafs = tree.Nodes[0]

	return index, leaf, leaf
}

func (tree *MerkleTree) getIntermediaryHashesByIndex(index int) (intermediaryHashes []*Node) {
	leafs := len(tree.Nodes[0])
	if index >= leafs {
		// TODO throw error
	}
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

func (tree *MerkleTree) IntermediaryHashesByIndex(index int) (intermediaryHashes []string) {
	hashes := tree.getIntermediaryHashesByIndex(index)
	intermediaryHashes = make([]string, len(hashes))
	for i, h := range hashes {
		intermediaryHashes[i] = h.Hash.Hex()
	}

	return intermediaryHashes
}

func (tree *MerkleTree) ValidateExistance(original []byte, index int, intermediaryHashes []string) bool {
	leafHash := crypto.Keccak256Hash(original)

	treeLeaf := tree.Nodes[0][index]

	if leafHash.Big().Cmp(treeLeaf.Hash.Big()) != 0 {
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

	return tempBHash.Big().Cmp(tree.Root.Hash.Big()) == 0

}

func New() *MerkleTree {
	var tree MerkleTree
	tree.init()

	return &tree
}
