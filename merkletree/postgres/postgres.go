package postgres

import (
	memory "tree/merkletree/memory"
)

type PostgresMerkleTree struct {
	*memory.MemoryMerkleTree
}

func (tree *PostgresMerkleTree) Add(data []byte) (index int, hash string) {
	// TODO add saving to database
	index, hash = tree.MemoryMerkleTree.Add(data)
	return index, hash
}

// NewTree returns a pointer to an initialized PostgresMerkleTree
func NewTree() *PostgresMerkleTree {

	memoryTree := memory.NewTree()

	postgresMemoryTree := PostgresMerkleTree{}

	postgresMemoryTree.MemoryMerkleTree = memoryTree

	return &postgresMemoryTree
}

// TODO load from database
func LoadTree() *PostgresMerkleTree {
	return NewTree()
}
