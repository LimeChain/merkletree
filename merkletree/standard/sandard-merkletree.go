package standard

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

type StandardNode struct {
	hash   common.Hash
	index  int
	parent *StandardNode
}

func (node *StandardNode) Hash() string {
	return node.hash.Hex()
}

func (n StandardNode) String() string {
	return n.hash.Hex()
}

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
	node := len(tree.nodes[level])
	if index >= node {
		// TODO throw error
	}
	if index%2 == 1 {
		return tree.nodes[level][index-1]
	}

	if index == node-1 {
		return tree.nodes[level][index]
	}

	return tree.nodes[level][index+1]
}

func (tree *StandardMerkleTree) getLeafSibling(index int) *StandardNode {
	return tree.getNodeSibling(0, index)
}

func (tree *StandardMerkleTree) getIntermediaryHashesByIndex(index int) (intermediaryHashes []*StandardNode) {
	leafs := len(tree.nodes[0])
	if index >= leafs {
		// TODO throw error
	}
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
	return index, leaf.String()
}

func (tree *StandardMerkleTree) IntermediaryHashesByIndex(index int) (intermediaryHashes []string) {
	hashes := tree.getIntermediaryHashesByIndex(index)
	intermediaryHashes = make([]string, len(hashes))
	for i, h := range hashes {
		intermediaryHashes[i] = h.hash.Hex()
	}

	return intermediaryHashes
}

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

func (tree *StandardMerkleTree) Root() string {
	return tree.root.String()
}

func (tree *StandardMerkleTree) Length() uint {
	return uint(len(tree.nodes[0]))
}

func (tree StandardMerkleTree) String() string {
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

func (tree *StandardMerkleTree) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("{\"root\":\"%v\", \"length\":%v}", tree.Root(), tree.Length())
	return []byte(res), nil
}

func New() *StandardMerkleTree {
	var tree StandardMerkleTree
	tree.init()

	return &tree
}
