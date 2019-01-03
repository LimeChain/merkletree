package postgres

import (
	"../../merkletree"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	InsertQuery       = "INSERT INTO hashes (hash) VALUES ($1)"
	SelectQuery       = "SELECT hash FROM hashes ORDER BY id"
	CreateQuery       = "CREATE TABLE hashes(id SERIAL PRIMARY KEY,hash VARCHAR(66) NOT NULL);"
	CreateIfNotExists = "CREATE TABLE IF NOT EXISTS hashes(id SERIAL PRIMARY KEY,hash VARCHAR(66) NOT NULL);"
)

type PostgresMerkleTree struct {
	merkletree.MarshalledMerkleTree
	db *sql.DB
}

func (tree *PostgresMerkleTree) Add(data []byte) (index int, hash string) {
	index, hash = tree.MarshalledMerkleTree.Add(data)
	_, err := tree.db.Exec(InsertQuery, hash)
	if err != nil {
		fmt.Println(err.Error())
	}
	return index, hash
}

// LoadMerkleTree takes an implementation of Merkle tree and postgre connection string
// Augments the tree with db saving
// returns a pointer to an initialized PostgresMerkleTree
func LoadMerkleTree(tree merkletree.MarshalledMerkleTree, connStr string) *PostgresMerkleTree {

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(CreateIfNotExists)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(SelectQuery)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var hash string
		err = rows.Scan(&hash)
		if err != nil {
			panic(err)
		}
		tree.Insert(hash)
	}

	postgresMemoryTree := PostgresMerkleTree{}
	postgresMemoryTree.db = db
	postgresMemoryTree.MarshalledMerkleTree = tree

	return &postgresMemoryTree
}
