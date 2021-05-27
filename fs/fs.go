package fs

type FileHeader struct {
	Name string
	Size, ModTime, Inode uint64
	State State
}

type DirHeader struct {
	Relpath, Name string
	ModTime, Inode uint64 // size = len(dh.Children)
	State State
	Children []FileHeader
}


type State uint16

const (
	Unchecked State = iota
	Unmodified
	Modified
	Deleted

	Ignored
)