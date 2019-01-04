package merkletree

import (
	"encoding/json"
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
	IntermediaryHashesByIndex(index int) (intermediaryHashes []string, err error)
	ValidateExistence(original []byte, index int, intermediaryHashes []string) (bool, error)
	HashAt(index int) (string, error)
	Root() string
	Length() int
}

type Internaler interface {
	Insert(hash string) (index int)
	// TODO Recalculate to be addet too
}

type InternalMerkleTree interface {
	MerkleTree
	Internaler
}

type Externaler interface {
	json.Marshaler
}

type ExternalMerkleTree interface {
	json.Marshaler
	MerkleTree
}

type FullMerkleTree interface {
	MerkleTree
	Internaler
	Externaler
}
