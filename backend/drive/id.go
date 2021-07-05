package drive

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/liamvdv/sharedHome/backend"
	. "github.com/liamvdv/sharedHome/util"
	"google.golang.org/api/drive/v3"
)

/*
	Our files need to have a parent, who's id can't be deduced from the filepath,
	since Google Drive is neither hierachical, nor ensures it name uniqueness.
	Files have to be linked to their parents, so we need to build a local
	representation of the remote file system.
	We will store the hash paths to a map, which we can then serialise to a cache.
*/

// remoteId is a workaround because the drive isn't hierarchical.
type remoteId struct {
	mu  sync.RWMutex
	ids map[string]string
}

// newRemoteId first looks if it can read from the provided cache. If it exists
// but reading from it fails, newRemoteId deletes it and moves on to the second option.
// If it cannot read from cache, newRemoteID fetches all directories from drive,
// and stores them in the cache. If it is not possible to fetch from remote,
// an error is returned and the application to be terminated.
func newRemoteId(cacheFp string, srv *drive.Service) *remoteId {
	r := remoteId{}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := remoteIdFromCache(cacheFp, r.ids); err == nil {
		return &r
	}
	if err := remoteIdFromRemote(srv, r.ids); err != nil {
		// TODO
		return nil
	}
	return &r
}

// remoteIdFromCache will try to load the cache located in fp to ids. If that is
// not possible because the cache doesn't exist, or the decoding fails, it returns
// a non-nil error. If the decoding fails, it will also delete the cache file.
func remoteIdFromCache(fp string, ids map[string]string) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	err = dec.Decode(&ids)
	if err != nil {
		log.Printf("gob: remoteId cache corrupted %v\n", err)
		if err := os.Remove(fp); err != nil {
			log.Printf("remoteId: cannot remove corrupted cache at %q %v\n", fp, err)
		}
		return err
	}
	return nil
}

type tempTree struct {
	name     string
	id       string
	parentId string
}

// remoteIdFromeRemote gets all folders in the drive and inserts them into the tempTree.
// Can use linear insertion because folders are ordered by mod time asc, so
// parent folders are guaratneed to come before child folders.
func remoteIdFromRemote(srv *drive.Service, ids map[string]string) error {
	q := fmt.Sprintf("trashed = false and mimeType = 'application/vnd.google-apps.folder'")
	o := "modifiedTime asc"

	var list *drive.FileList
	var err error
	var nextPageToken string
	for more := true; more; more = nextPageToken != "" {
		req := srv.Files.List().Fields("nextPageToken, files(id, name, parents)").Q(q)
		if nextPageToken != "" {
			req = req.PageToken(nextPageToken)
		}
		list, err = req.OrderBy(o).Do()
		if err != nil {
			return err
		}
		if len(list.Files) == 0 {
			log.Println("no remote files found.")
		}
		// linear search fine because ordered and remote folder last mtime 1970.
		// BUG(liamvdv): Actually, this is only true for the created time, which we don't know for the local files.
		seen := make(map[string]string)
		for _, f := range list.Files {
			log.Printf("Name: %s Id: %s Parents: %v\n", f.Name, f.Id, f.Parents)
			// key: id value: relpath, backend.RemoteFolderName has relpath ""
			// check if a parent exists and if so, add it.
			if dp, exists := seen[f.Parents[0]]; exists {
				relpath := filepath.Join(dp, f.Name)
				ids[relpath] = f.Id
				seen[f.Id] = relpath
			} else if f.Name == backend.RemoteFolderName {
				seen[f.Id] = ""
			}
		}
		nextPageToken = list.NextPageToken
	}
	return nil
}

// TODO(liamvdv): remove TestRemoveIdList
func TestRemoteIdList() {
	// util.PrepareTestConfigInit()
	// config.Init()

	drive, err := New()
	if err != nil {
		log.Println(err)
		return
	}
	// fp := filepath.Join(config.Temp, "drive-id-cache.gob")

	m := make(map[string]string)
	if err := remoteIdFromRemote(drive.srv, m); err != nil {
		log.Println(err)
	}
}

func remoteIdToCache(fp string, ids map[string]string) error {
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer SaveClose(f)
	enc := gob.NewEncoder(f)
	return enc.Encode(&ids)
}

func (r *remoteId) get(relpath string) string {
	r.mu.RLock()
	ret := r.ids[relpath]
	r.mu.RUnlock()
	return ret
}
