package merkletree

type Node interface {
	Hash() string
	String() string
}
