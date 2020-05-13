package post

type SortType string

const (
	FLAT        SortType = "flat"
	TREE        SortType = "tree"
	PARENT_TREE SortType = "parent_tree"
	DEFAULT     SortType = "default"
)
