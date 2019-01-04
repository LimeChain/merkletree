package baseapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/LimeChain/merkle-tree-api/merkletree/memory"
	"github.com/LimeChain/merkle-tree-api/merkletree/merkletreetest"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertValidResponse(et *merkletreetest.ExtendedTesting, resp *http.Response, err error) {
	et.Assert(err == nil, "Error was thrown by the API on Request")
	et.Assert(resp != nil, "Response was not returned")
}

func TestMerkleTreeStatus(t *testing.T) {
	et := merkletreetest.WrapTesting(t)
	tree := memory.NewMerkleTree()

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
		treeRouter = MerkleTreeStatus(treeRouter, tree)
		r.Mount("/api/merkletree", treeRouter)
	})
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/v1/api/merkletree")
	assertValidResponse(et, resp, err)

	decoder := json.NewDecoder(resp.Body)
	var r treeStatusResponse
	err = decoder.Decode(&r)
	et.Assert(err == nil, "Error was thrown when parsing the response")
	et.Assert(r.Status, "The status for getting the tree status was false")
	et.Assert(r.Tree == nil, "The tree was not nil on status before insertion")

	data1 := []byte("First Leaf")
	_, h := tree.Add(data1)

	resp, err = server.Client().Get(server.URL + "/v1/api/merkletree")
	assertValidResponse(et, resp, err)

	body, err := ioutil.ReadAll(resp.Body)
	et.Assert(err == nil, "Could not read the response body")
	expected := fmt.Sprintf(`{"status":true,"tree":{"root":"%v","length":%v}}`, h, 1)
	et.Assert(strings.TrimSpace(string(body)) == expected, "The response was not the expected one")
}

func TestMerkleTreeInsert(t *testing.T) {
	et := merkletreetest.WrapTesting(t)
	tree := memory.NewMerkleTree()

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
		treeRouter = MerkleTreeInsert(treeRouter, tree)
		r.Mount("/api/merkletree", treeRouter)
	})
	server := httptest.NewServer(router)
	defer server.Close()

	data := "TestData"

	data1 := []byte(data)
	expectedHash := crypto.Keccak256Hash(data1).Hex()

	req := addDataRequest{
		Data: data,
	}

	reqString, err := json.Marshal(req)

	resp, err := server.Client().Post(server.URL+"/v1/api/merkletree", "application/json", bytes.NewBuffer(reqString))
	assertValidResponse(et, resp, err)

	decoder := json.NewDecoder(resp.Body)
	var r addDataResponse
	err = decoder.Decode(&r)
	et.Assert(err == nil, "Error was thrown when parsing the response")
	et.Assert(r.Status, "The status for inserting in the tree status was false")
	et.Assert(r.Index == 0, "The inserted index was not 0 for addition")
	et.Assert(r.Hash == expectedHash, "The returned hash was not the hash of the passed data")

	type fakeStruct struct {
		b float64
	}

	fs1 := fakeStruct{14.6}

	reqString, err = json.Marshal(fs1)

	resp, err = server.Client().Post(server.URL+"/v1/api/merkletree", "application/json", bytes.NewBuffer(reqString))

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	et.Assert(err == nil, "Error was thrown when parsing the response")
	et.Assert(!r.Status, "The status for wrong data insert was true")
	et.Assert(r.Error == "Missing data field", "The status for wrong data insert was true")
	et.Assert(r.Index == -1, "The inserted index was not -1 for wrong addition")

	resp, err = server.Client().Post(server.URL+"/v1/api/merkletree", "application/json", bytes.NewBufferString("wrong-json-format"))

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	et.Assert(err == nil, "Error was thrown when parsing the response")
	et.Assert(!r.Status, "The status for wrong data insert was true")
	et.Assert(r.Index == -1, "The inserted index was not -1 for wrong addition")

}
