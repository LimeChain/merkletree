package merkletree

type MerkleTree interface {
	Add(data []byte) (index int, hash string)
	IntermediaryHashesByIndex(index int) (intermediaryHashes []string)
	ValidateExistence(original []byte, index int, intermediaryHashes []string) bool
	Root() string
	Length() uint
	String() string
}
