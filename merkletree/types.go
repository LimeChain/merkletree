package merkletree

import (
	"fmt"
)

// Node represents a single node in a Merkle tree
type Node interface {
	fmt.Stringer
	Hash() string
	Index() int
}

// MerkleTree defines and represents the methods a generic Merkle tree should have
type MerkleTree interface {
	fmt.Stringer
	Add(data []byte) (index int, hash string)
	IntermediaryHashesByIndex(index int) (intermediaryHashes []string)
	ValidateExistence(original []byte, index int, intermediaryHashes []string) bool
	HashAt(index int) string
	Root() string
	Length() int
}
