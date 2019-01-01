package baseapi

import (
	"../../../merkletree"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"io/ioutil"
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

type merkleAPIResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error,omitempty"`
}

type treeStatusResponse struct {
	merkleAPIResponse
	Tree merkletree.MerkleTree `json:"tree"`
}

func getTreeStatus(tree merkletree.MarshalledMerkleTree) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if tree.Length() == 0 {
			render.JSON(w, r, treeStatusResponse{merkleAPIResponse{true, ""}, nil})
			return
		}
		render.JSON(w, r, treeStatusResponse{merkleAPIResponse{true, ""}, tree})
		return
	}
	return handler
}

type intermediaryHashesResponse struct {
	merkleAPIResponse
	Hashes []string `json:"hashes,omitempty"`
}

func getIntermediaryHashesHandler(tree merkletree.MerkleTree) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		index, err := strconv.Atoi(chi.URLParam(r, "index"))
		if err != nil {
			render.JSON(w, r, intermediaryHashesResponse{merkleAPIResponse{false, err.Error()}, nil})
			return
		}
		hashes, err := tree.IntermediaryHashesByIndex(index)
		if err != nil {
			render.JSON(w, r, intermediaryHashesResponse{merkleAPIResponse{false, err.Error()}, nil})
			return
		}
		render.JSON(w, r, intermediaryHashesResponse{merkleAPIResponse{true, ""}, hashes})
	}
	return handler
}

type addDataResponse struct {
	merkleAPIResponse
	Index int    `json:"index"`
	Hash  string `json:"hash,omitempty"`
}

func addDataHandler(tree merkletree.MerkleTree) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			render.JSON(w, r, addDataResponse{merkleAPIResponse{false, err.Error()}, -1, ""})
			return
		}
		index, hash := tree.Add(data)
		render.JSON(w, r, addDataResponse{merkleAPIResponse{true, ""}, index, hash})
		// log.Println(tree)
	}
	return handler
}
