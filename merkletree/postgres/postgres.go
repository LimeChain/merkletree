package postgres

import (
	"../../merkletree"
)

type PostgresMerkleTree struct {
	merkletree.MarshalledMerkleTree
}

func (tree *PostgresMerkleTree) Add(data []byte) (index int, hash string) {
	// TODO add saving to database
	index, hash = tree.MarshalledMerkleTree.Add(data)
	return index, hash
}

// NewMerkleTree takes an implementation of Merkle tree and augments it with saving data to postgres database
// returns a pointer to an initialized PostgresMerkleTree
func NewMerkleTree(tree merkletree.MarshalledMerkleTree) *PostgresMerkleTree {

	postgresMemoryTree := PostgresMerkleTree{}

	postgresMemoryTree.MarshalledMerkleTree = tree

	return &postgresMemoryTree
}

func LoadMerkleTree(tree merkletree.MarshalledMerkleTree) *PostgresMerkleTree {
	// TODO load from database
	return NewMerkleTree(tree)
}
