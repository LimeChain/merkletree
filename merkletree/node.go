package merkletree

type Node interface {
	Hash() string
	Index() int
	String() string
}
