package main

import (
	"./merkletree"
	"./merkletree/memory"
	"./merkletree/postgres"
	merkleRestAPI "./merkletree/restapi/baseapi"
	validateAPI "./merkletree/restapi/validateapi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log"
	"net/http"
)

func createAndStartAPI(tree merkletree.ExternalMerkleTree) {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
	)

	router.Route("/v1", func(r chi.Router) {
		treeRouter := chi.NewRouter()
		treeRouter = merkleRestAPI.MerkleTreeStatus(treeRouter, tree)
		treeRouter = merkleRestAPI.MerkleTreeInsert(treeRouter, tree)
		treeRouter = merkleRestAPI.MerkleTreeHashes(treeRouter, tree)
		treeRouter = validateAPI.MerkleTreeValidate(treeRouter, tree)
		r.Mount("/api/merkletree", treeRouter)
	})
	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	// elements := 1000000
	connStr := "user=georgespasov dbname=postgres port=5432 sslmode=disable"
	tree := postgres.LoadMerkleTree(memory.NewMerkleTree(), connStr)
	// tree := postgres.NewMerkleTree(memory.NewMerkleTree(), connStr)
	// for i := 0; i < elements; i++ {
	// 	tree.Add([]byte("hello" + strconv.Itoa(i)))
	// }

	// tree.Add(make([]byte, 1024*1024))

	// rawData := []byte("Ogi e Majstor")
	// index, _ := tree.Add(rawData)
	// intermediaryHashes, err := tree.IntermediaryHashesByIndex(index)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// s, _ := tree.ValidateExistence(rawData, index, intermediaryHashes)
	// fmt.Printf("%v exists in the tree: %v\n", string(rawData), s)
	// s, _ = tree.ValidateExistence(rawData[:7], index, intermediaryHashes)
	// fmt.Printf("%v exists in the tree: %v\n", string(rawData[:7]), s)

	// bs, err := json.Marshal(tree)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(bs))

	createAndStartAPI(tree)

}
