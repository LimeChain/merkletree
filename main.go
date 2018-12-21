package main

import (
	"encoding/json"
	"fmt"
	"tree/merkletree"
)

func main() {
	elements := 25
	tree := merkletree.New()
	for i := 0; i < elements; i++ {
		tree.Add([]byte("hello" + string(i)))
	}

	tree.Add(make([]byte, 1024*1024))

	rawData := []byte("Ogi e Majstor")
	index, _ := tree.Add(rawData)
	intermediaryHashes := tree.IntermediaryHashesByIndex(index)

	fmt.Println(tree)
	fmt.Printf("%v exists in the tree: %v\n", string(rawData), tree.ValidateExistance(rawData, index, intermediaryHashes))
	fmt.Printf("%v exists in the tree: %v\n", string(rawData[:7]), tree.ValidateExistance(rawData[:7], index, intermediaryHashes))

	bs, err := json.Marshal(tree)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(bs))

}
