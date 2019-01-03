package baseapi

import (
	"../../../merkletree"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"net/http"
	"strconv"
)

// MerkleTreeStatus takes pointer to initialized router and the merkle tree and exposes Rest API routes for getting of status
func MerkleTreeStatus(treeRouter *chi.Mux, tree merkletree.MarshalledMerkleTree) *chi.Mux {
	treeRouter.Get("/", getTreeStatus(tree))
	return treeRouter
}

// MerkleTreeHashes takes pointer to initialized router and the merkle tree and exposes Rest API routes for getting of intermediary hashes
func MerkleTreeHashes(treeRouter *chi.Mux, tree merkletree.MerkleTree) *chi.Mux {
	treeRouter.Get("/hashes/{index}", getIntermediaryHashesHandler(tree))
	return treeRouter
}

// MerkleTreeInsert takes pointer to initialized router and the merkle tree and exposes Rest API routes for addition
func MerkleTreeInsert(treeRouter *chi.Mux, tree merkletree.MerkleTree) *chi.Mux {
	treeRouter.Post("/", addDataHandler(tree))
	return treeRouter
}

// MerkleAPIResponse represents the minimal response structure
type MerkleAPIResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error,omitempty"`
}

type treeStatusResponse struct {
	MerkleAPIResponse
	Tree merkletree.MerkleTree `json:"tree"`
}

func getTreeStatus(tree merkletree.MarshalledMerkleTree) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if tree.Length() == 0 {
			render.JSON(w, r, treeStatusResponse{MerkleAPIResponse{true, ""}, nil})
			return
		}
		render.JSON(w, r, treeStatusResponse{MerkleAPIResponse{true, ""}, tree})
		return
	}
	return handler
}

type intermediaryHashesResponse struct {
	MerkleAPIResponse
	Hashes []string `json:"hashes,omitempty"`
}

func getIntermediaryHashesHandler(tree merkletree.MerkleTree) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		index, err := strconv.Atoi(chi.URLParam(r, "index"))
		if err != nil {
			render.JSON(w, r, intermediaryHashesResponse{MerkleAPIResponse{false, err.Error()}, nil})
			return
		}
		hashes, err := tree.IntermediaryHashesByIndex(index)
		if err != nil {
			render.JSON(w, r, intermediaryHashesResponse{MerkleAPIResponse{false, err.Error()}, nil})
			return
		}
		render.JSON(w, r, intermediaryHashesResponse{MerkleAPIResponse{true, ""}, hashes})
	}
	return handler
}

type addDataRequest struct {
	Data string `json:"data"`
}

type addDataResponse struct {
	MerkleAPIResponse
	Index int    `json:"index"`
	Hash  string `json:"hash,omitempty"`
}

func addDataHandler(tree merkletree.MerkleTree) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var b addDataRequest
		err := decoder.Decode(&b)
		if err != nil {
			render.JSON(w, r, addDataResponse{MerkleAPIResponse{false, err.Error()}, -1, ""})
			return
		}

		if b.Data == "" {
			render.JSON(w, r, addDataResponse{MerkleAPIResponse{false, "Missing data field"}, -1, ""})
			return
		}
		fmt.Println(b.Data)
		index, hash := tree.Add([]byte(b.Data))
		render.JSON(w, r, addDataResponse{MerkleAPIResponse{true, ""}, index, hash})
	}
	return handler
}
