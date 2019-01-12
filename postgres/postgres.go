package postgres

import (
	"database/sql"
	"fmt"
	"github.com/LimeChain/merkletree"
	_ "github.com/lib/pq"
	"sync"
)

const (
	InsertQuery       = "INSERT INTO hashes (hash) VALUES ($1)"
	SelectQuery       = "SELECT hash FROM hashes ORDER BY id"
	CreateQuery       = "CREATE TABLE hashes(id SERIAL PRIMARY KEY,hash VARCHAR(66) NOT NULL);"
	CreateIfNotExists = "CREATE TABLE IF NOT EXISTS hashes(id SERIAL PRIMARY KEY,hash VARCHAR(66) NOT NULL);"
)

type PostgresMerkleTree struct {
	merkletree.FullMerkleTree
	db    *sql.DB
	mutex sync.Mutex
}

func (tree *PostgresMerkleTree) Add(data []byte) (index int, hash string) {
	tree.mutex.Lock()
	index, hash = tree.FullMerkleTree.Add(data)
	tree.addHashToDB(hash)
	tree.mutex.Unlock()
	return index, hash
}

func (tree *PostgresMerkleTree) addHashToDB(hash string) {
	_, err := tree.db.Exec(InsertQuery, hash)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func connectToDb(connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic("Could not connect to the database.\n Original error: " + err.Error())
	}
	return db
}

func createHashesTable(db *sql.DB) {
	_, err := db.Exec(CreateIfNotExists)
	if err != nil {
		panic("Could not create the table in the db.\n Original error: " + err.Error())
	}
}

func getAndInsertStoredHashes(db *sql.DB, tree merkletree.InternalMerkleTree) {
	rows, err := db.Query(SelectQuery)
	if err != nil {
		panic("Could not query the stored hashes.\n Original error: " + err.Error())
	}

	for rows.Next() {
		var hash string
		err = rows.Scan(&hash)
		if err != nil {
			panic("Could not scan the stored hashes.\n Original error: " + err.Error())
		}
		tree.RawInsert(hash)
	}

	tree.Recalculate()
}

// LoadMerkleTree takes an implementation of Merkle tree and postgre connection string
// Augments the tree with db saving
// returns a pointer to an initialized PostgresMerkleTree
func LoadMerkleTree(tree merkletree.FullMerkleTree, connStr string) *PostgresMerkleTree {

	db := connectToDb(connStr)

	createHashesTable(db)

	getAndInsertStoredHashes(db, tree)

	postgresMemoryTree := PostgresMerkleTree{}
	postgresMemoryTree.db = db
	postgresMemoryTree.FullMerkleTree = tree

	return &postgresMemoryTree
}
