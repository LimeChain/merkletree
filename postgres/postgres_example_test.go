package postgres_test

import (
	"github.com/LimeChain/merkletree/memory"
	"github.com/LimeChain/merkletree/postgres"
)

func Example() {
	connStr := "user=merkle dbname=merrymerkle port=54321 sslmode=disable"
	tree := postgres.LoadMerkleTree(memory.NewMerkleTree(), connStr)
	data := "Merkle Trees Rock"
	tree.Add([]byte(data))
}
