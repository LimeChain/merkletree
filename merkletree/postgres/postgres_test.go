package postgres

import (
	"../../merkletree"
	"../../merkletree/memory"
	"testing"
)

func assert(t *testing.T, condition bool, msg string) {

	if !condition {
		t.Error(msg)
	}
}

func TestNewMerkleTree(t *testing.T) {
	underlyingTree := memory.NewMerkleTree()
	tree := NewMerkleTree(underlyingTree)
	assert(t, tree.MarshalledMerkleTree == underlyingTree, "The underlying tree was not set correctly")
	_, isMerkleTree := interface{}(tree).(merkletree.MerkleTree)
	assert(t, isMerkleTree, "The tree did not implement the MerkleTree interface")
	_, isMarshalledMerkleTree := interface{}(tree).(merkletree.MarshalledMerkleTree)
	assert(t, isMarshalledMerkleTree, "The tree did not implement the MarshalledMerkleTree interface")
}
