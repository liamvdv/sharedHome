package fs

type FileHeader struct {
	Name string
	Size, ModTime, Inode uint64
	State State
}

type DirHeader struct {
	Relpath, Name string
	State State
	Children []FileHeader // size = len(dh.Children)
}


type State uint16

const (
	Unchecked State = iota
	Unmodified
	Modified
	Deleted

	Ignored
)