package core

import (
	"path"

	"github.com/liamvdv/sharedHome/vfs"
)

// core should later implement the concurrent index tranversal and comparsion,
// as well as the calls into the different libraries  to fulfill the tasks.

type Sync struct {
	local   vfs.FileIndex
	remote  vfs.FileIndex
	changed map[string]struct {
		target *vfs.File
		update vfs.File
	}
	indexUpdateRequired bool
}

type Task interface {
	IsNetworkBound() bool
}

type Delete struct {}

type Download struct{}

type Upload struct{}

type MetadataChangeLocal struct{}

// When I speak of downloading a file I mean that real files should be downloaded
// and directories should be created locally.

// should limit number of CompareDirs with semaphore pattern (chan buff size is max active CompareDir.)
// One of local or remote must be non-nil.
func (s *Sync) compareDir(local, remote *vfs.File, task chan<- Task) {

	if local == nil {
		if remote.State == vfs.Ignored || remote.State == vfs.Deleted {
			return
		}
		// either download file or file was deleted locally.
		rdp := path.Dir(remote.Relpath)
		lpar, err := s.local.GetDir(rdp)
		if err != nil {
			panic(err) // what to do?
		}
		switch {
		// can compare parent to child since child must have child.MTime >= par.MTime because child.MTime = child.CTime when created.
		case lpar.MTime > remote.MTime:
			// file was deleted locally because local dir changes more current.
			// not 100% sure since other changes may have also triggered that.....
			// change state of File and place in index. and mark that index upload is required

		case lpar.MTime < remote.MTime:
			// file needs to be downloaded; 100% sure
			// recursively download!
		case lpar.MTime == remote.MTime:
			// don't know how this is possible
			panic("invalid state")
		}
		return
	}

	if remote == nil {
		if local.State == vfs.Ignored { // || local.State == vfs.Deleted not possible since it wouldn't be in the index then.
			// store this change to the index and mark that index upload is required.
			return
		}
		// either upload dir recursively or dir was deleted remotely.
		ldp := path.Dir(local.Relpath)
		rpar, err := s.remote.GetDir(ldp)
		if err != nil {
			panic(err) // what to do?
		}
		switch {
		case local.MTime > rpar.MTime:
			// file was created locally because local dir changes are younger.
			// 100% sure
			// file needs to be uploaded -> recursively!
		case local.MTime < rpar.MTime:
			// file was deleted remotely, not 100% sure since other changes might have happend in dir.
		case local.MTime == rpar.MTime:
			// don't know how this is possible
			panic("invalid state")
		}
		return
	}

	// compare remote and local dirs
	// ...

	// compare children -> lexcially ordered children would help a lot.
	// ....
}
