package memory_test

import (
	"fmt"
	"github.com/LimeChain/merkletree/memory"
	"strconv"
)

func Example() {
	tree := memory.NewMerkleTree()
	for i := 0; i < 10; i++ {
		tree.Add([]byte("hello" + strconv.Itoa(i)))
	}

	data := "Merkle Trees Rock"
	i, h := tree.Add([]byte(data))
	fmt.Printf("Index: %v\n", i)
	fmt.Printf("Hash: %v\n", h)
	intermediaryHashes, _ := tree.IntermediaryHashesByIndex(i)
	fmt.Printf("Hashes: %v\n", intermediaryHashes)
	exists, _ := tree.ValidateExistence([]byte(data), i, intermediaryHashes)
	fmt.Printf("Element Exists: %v\n", exists)
	fmt.Printf("Root: %v\n", tree.Root())

	// Output:
	// Index: 10
	// Hash: 0x3c8fcd181200324dec9b2ce359583f07257388a4ffcf4daa57ec11215e0a30d5
	// Hashes: [0x3c8fcd181200324dec9b2ce359583f07257388a4ffcf4daa57ec11215e0a30d5 0xb81ce7f7f1167b6dfc5d973978d8eee5a9e800814521179d70bc98c057df6422 0x72be81d1d885b8fc3215a13de267252b5dc5b910fddf409a4cb8bc651c17fc96 0x01f628f3ed77668f216c29e51412769a7d7e8819ca5e222d16918b149d8b6ecc]
	// Element Exists: true
	// Root: 0x43248fc6531118c33d162971d0b2587ce7528348d594c3d533df8b4c7f02f703
}
