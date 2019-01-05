package validateapi

import (
	"encoding/json"
	"fmt"
	"github.com/LimeChain/merkletree"
	"github.com/LimeChain/merkletree/restapi/baseapi"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"net/http"
)

// MerkleTreeValidate takes pointer to initialized router and the merkle tree and exposes Rest API routes for getting of status
func MerkleTreeValidate(treeRouter *chi.Mux, tree merkletree.ExternalMerkleTree) *chi.Mux {
	treeRouter.Post("/validate", validate(tree))
	return treeRouter
}

type validateRequest struct {
	Data   string   `json:"data"`
	Index  int      `json:"index"`
	Hashes []string `json:"hashes"`
}

type validateResponse struct {
	baseapi.MerkleAPIResponse
	Exists bool `json:"exists"`
}

func validate(tree merkletree.ExternalMerkleTree) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var b validateRequest
		err := decoder.Decode(&b)
		if err != nil {
			render.JSON(w, r, validateResponse{baseapi.MerkleAPIResponse{Status: false, Error: err.Error()}, false})
			return
		}

		if b.Data == "" {
			render.JSON(w, r, validateResponse{baseapi.MerkleAPIResponse{Status: false, Error: "Missing data field"}, false})
			return
		}
		fmt.Println(b.Data)
		fmt.Println(b.Index)
		exists, err := tree.ValidateExistence([]byte(b.Data), b.Index, b.Hashes)
		if err != nil {
			render.JSON(w, r, validateResponse{baseapi.MerkleAPIResponse{Status: false, Error: err.Error()}, false})
			return
		}

		render.JSON(w, r, validateResponse{baseapi.MerkleAPIResponse{Status: true, Error: ""}, exists})
	}
	return handler
}
