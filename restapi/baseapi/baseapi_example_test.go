package baseapi_test

import (
	"github.com/LimeChain/merkletree/memory"
	"github.com/LimeChain/merkletree/restapi/baseapi"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

func Example() {
	tree := memory.NewMerkleTree()
	router := chi.NewRouter()
	router.Route("/v1", func(r chi.Router) {
		treeRouter := chi.NewRouter()
		treeRouter = baseapi.MerkleTreeStatus(treeRouter, tree)
		treeRouter = baseapi.MerkleTreeInsert(treeRouter, tree)
		treeRouter = baseapi.MerkleTreeHashes(treeRouter, tree)
		r.Mount("/api/merkletree", treeRouter)
	})
	log.Fatal(http.ListenAndServe(":8080", router))
}
