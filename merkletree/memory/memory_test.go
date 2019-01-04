package memory

import (
	"../../merkletree"
	"../../merkletree/merkletreetest"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func TestNewMerkleTree(t *testing.T) {
	et := merkletreetest.WrapTesting(t)
	tree := NewMerkleTree()
	et.Assert(tree.RootNode == nil, "The tree root was not nil after init")
	et.Assert(len(tree.Nodes) == 1, "The tree did not have 1 initial level")
	_, isMerkleTree := interface{}(tree).(merkletree.MerkleTree)
	et.Assert(isMerkleTree, "The tree did not implement the MerkleTree interface")
	_, isExternalMerkleTree := interface{}(tree).(merkletree.ExternalMerkleTree)
	et.Assert(isExternalMerkleTree, "The tree did not implement the MarshalledMerkleTree interface")
	_, isInternalMerkleTree := interface{}(tree).(merkletree.InternalMerkleTree)
	et.Assert(isInternalMerkleTree, "The tree did not implement the InternalMerkleTree interface")
	_, isFullMerkleTree := interface{}(tree).(merkletree.FullMerkleTree)
	et.Assert(isFullMerkleTree, "The tree did not implement the FullMerkleTree interface")
}

func TestAdd(t *testing.T) {
	et := merkletreetest.WrapTesting(t)
	tree := NewMerkleTree()

	data1 := []byte("First Leaf")
	dh1 := crypto.Keccak256Hash(data1)
	expectedHash := dh1.Hex()
	i, h := tree.Add(data1)

	et.Assert(len(dh1[:]) == 32, "The hash should be 32 bytes")
	et.Assert(i == 0, "The index of first addition was not 0")
	et.Assert(h == expectedHash, "The hash of the added node was not the keccak256 hash of the data")
	et.Assert(tree.Root() == expectedHash, "The hash of the root was not equal to the only element")

	data2 := []byte("Second Leaf")
	dh2 := crypto.Keccak256Hash(data2)
	expectedHash = dh2.Hex()
	i, h = tree.Add(data2)
	expectedRoot := crypto.Keccak256Hash(dh1[:], dh2[:]).Hex()

	et.Assert(i == 1, "The index of second addition was not 1")
	et.Assert(h == expectedHash, "The hash of the added node was not the keccak256 hash of the data")

	et.Assert(tree.Root() == expectedRoot, "The hash of the root was not the hash of the two elements")
	et.Assert(len(tree.Nodes) == 2, "The tree did not grow to 2 levels")

	data3 := []byte("Third Leaf")
	dh3 := crypto.Keccak256Hash(data3)
	expectedHash = dh3.Hex()
	i, h = tree.Add(data3)
	leftBranch := crypto.Keccak256Hash(dh1[:], dh2[:])
	rightBranch := crypto.Keccak256Hash(dh3[:], dh3[:])

	expectedRoot = crypto.Keccak256Hash(leftBranch[:], rightBranch[:]).Hex()

	et.Assert(i == 2, "The index of third addition was not 2")
	et.Assert(h == expectedHash, "The hash of the added node was not the keccak256 hash of the data")

	et.Assert(tree.Root() == expectedRoot, "The hash of the root was not correctly calculated", tree)
	et.Assert(len(tree.Nodes) == 3, "The tree did not grow to 2 levels")

	data4 := []byte("Fourth Leaf")
	dh4 := crypto.Keccak256Hash(data4)
	expectedHash = dh4.Hex()
	i, h = tree.Add(data4)

	leftBranch = crypto.Keccak256Hash(dh1[:], dh2[:])
	rightBranch = crypto.Keccak256Hash(dh3[:], dh4[:])

	expectedRoot = crypto.Keccak256Hash(leftBranch[:], rightBranch[:]).Hex()

	et.Assert(i == 3, "The index of third addition was not 2")
	et.Assert(h == expectedHash, "The hash of the added node was not the keccak256 hash of the data")

	et.Assert(tree.Root() == expectedRoot, "The hash of the root was not correctly calculated", tree)
	et.Assert(len(tree.Nodes) == 3, "The tree was not 2 levels after fourth addition")
}

func TestIntermediaryHashesByIndex(t *testing.T) {
	et := merkletreetest.WrapTesting(t)
	tree := NewMerkleTree()

	_, err := tree.IntermediaryHashesByIndex(1)

	et.Assert(err != nil, "Error was not thrown")
	et.Assert(err.Error() == outOfBounds, "Incorrect message was thrown on fetching hashes by out of bounds index")

	data1 := []byte("First Leaf")
	dh1 := crypto.Keccak256Hash(data1)

	tree.Add(data1)

	hashes, err := tree.IntermediaryHashesByIndex(0)

	et.Assert(err == nil, "Error was thrown")
	et.Assert(len(hashes) == 0, "Too many intermediary hashes were returned")

	data2 := []byte("Second Leaf")
	dh2 := crypto.Keccak256Hash(data2)
	tree.Add(data2)

	data3 := []byte("Third Leaf")
	dh3 := crypto.Keccak256Hash(data3)
	tree.Add(data3)

	expectedHash0 := dh1.Hex()
	expectedHash1 := crypto.Keccak256Hash(dh3[:], dh3[:]).Hex()

	hashes, err = tree.IntermediaryHashesByIndex(1)

	et.Assert(err == nil, "Error was thrown")
	et.Assert(len(hashes) == 2, "Too many intermediary hashes were returned")
	et.Assert(hashes[0] == expectedHash0, "Incorrect intermediary hash at level 0 "+expectedHash0)
	et.Assert(hashes[1] == expectedHash1, "Incorrect intermediary hash at level 1 "+expectedHash0)

	expectedHash0 = dh3.Hex()
	expectedHash1 = crypto.Keccak256Hash(dh1[:], dh2[:]).Hex()

	hashes, err = tree.IntermediaryHashesByIndex(2)

	et.Assert(err == nil, "Error was thrown")
	et.Assert(len(hashes) == 2, "Too many intermediary hashes were returned")
	et.Assert(hashes[0] == expectedHash0, "Incorrect intermediary hash at level 0 "+expectedHash0)
	et.Assert(hashes[1] == expectedHash1, "Incorrect intermediary hash at level 1 "+expectedHash0)

}

func TestValidateExistence(t *testing.T) {
	et := merkletreetest.WrapTesting(t)

	tree := NewMerkleTree()

	data1 := []byte("First Leaf")
	tree.Add(data1)

	data2 := []byte("Second Leaf")
	tree.Add(data2)

	data3 := []byte("Third Leaf")
	tree.Add(data3)

	hashes, err := tree.IntermediaryHashesByIndex(1)

	et.Assert(err == nil, "Error was thrown for intermediary hashes")

	result, err := tree.ValidateExistence(data2, 1, hashes)

	et.Assert(err == nil, "Error was thrown on validating data2")
	et.Assert(result, "Did not find the original data on index 1")

	result, err = tree.ValidateExistence(data1, 1, hashes)

	et.Assert(err == nil, "Error was thrown on validating data1")
	et.Assert(!result, "Found correct data1 on index 1 but should not have found one")

	result, err = tree.ValidateExistence(data2, 2, hashes)

	et.Assert(err == nil, "Error was thrown on validating data2")
	et.Assert(!result, "Found correct data1 on index 2 but should not have found one")

	result, err = tree.ValidateExistence(data2, 1, []string{})

	et.Assert(err == nil, "Error was thrown on validating data2")
	et.Assert(!result, "Found correct data2 on index 1 without intermediary hashes but should not have found one")

	result, err = tree.ValidateExistence(data2, 10, hashes)

	et.Assert(err != nil, "Error was thrown on index out of boundes")
	et.Assert(err.Error() == outOfBounds, "Incorrect message was thrown on validating out of bounds index")

}

func TestLength(t *testing.T) {
	et := merkletreetest.WrapTesting(t)

	tree := NewMerkleTree()

	data1 := []byte("First Leaf")
	tree.Add(data1)

	et.Assert(tree.Length() == 1, "The length of the tree was not 1 after 1 addition")

	data2 := []byte("Second Leaf")
	tree.Add(data2)

	et.Assert(tree.Length() == 2, "The length of the tree was not 2 after second")

}

func TestString(t *testing.T) {
	et := merkletreetest.WrapTesting(t)

	tree := NewMerkleTree()

	data1 := []byte("First Leaf")
	tree.Add(data1)

	data2 := []byte("Second Leaf")
	tree.Add(data2)

	expected := `Level: 1, Count: 1
0x079c36e0e4573fd7169dfb6f6397bea69db51ca66bce0299a0ec643bd5996721	
Level: 0, Count: 2
0x1d47c3db13342a2ebb20fd47631d12370cf0dd323fe39503597bf38bc107cd5f	0x87f67c448ec13f7c376de9f12b39ae06a3b8d7c64ee9e311b30d428a5ed9fe91	
`

	received := tree.String()

	fmt.Println(expected)
	fmt.Println(received)

	et.Assert(received == expected, tree.String())

}

func TestHashAt(t *testing.T) {
	et := merkletreetest.WrapTesting(t)

	tree := NewMerkleTree()

	data1 := []byte("First Leaf")
	dh1 := crypto.Keccak256Hash(data1)
	tree.Add(data1)

	hash, err := tree.HashAt(0)

	et.Assert(err == nil, "Error was thrown for fetching first hash")
	et.Assert(hash == dh1.Hex(), "Incorrect hash was returned")

	_, err = tree.HashAt(5)

	et.Assert(err != nil, "Error was thrown on index out of boundes")
	et.Assert(err.Error() == outOfBounds, "Incorrect message was thrown on requesting hash at index out of bounds")

}

func TestMarshalJSON(t *testing.T) {
	et := merkletreetest.WrapTesting(t)

	tree := NewMerkleTree()

	data1 := []byte("First Leaf")
	tree.Add(data1)

	data2 := []byte("Second Leaf")
	tree.Add(data2)

	expected := `{"root":"0x079c36e0e4573fd7169dfb6f6397bea69db51ca66bce0299a0ec643bd5996721", "length":2}`

	received, _ := tree.MarshalJSON()

	et.Assert(string(received) == expected, string(received))
}

func TestNode(t *testing.T) {
	et := merkletreetest.WrapTesting(t)
	data := "TestData"
	h := crypto.Keccak256Hash([]byte(data))
	n := Node{
		hash:   h,
		Parent: nil,
		index:  5, // Parent index is always the current node index divided by two
	}

	et.Assert(n.Index() == 5, "The index returned was not correct")
	et.Assert(n.Hash() == h.Hex(), "The hash returned was not correct")
	et.Assert(n.String() == h.Hex(), "The hash returned was not correct")
}

func BenchmarkAdd(b *testing.B) {
	tree := NewMerkleTree()

	for i := 0; i < 1000000; i++ {
		tree.Add([]byte("First Leaf" + string(i)))
	}
}
