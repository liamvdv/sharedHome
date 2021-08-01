package remote

import (
	"context"

	"github.com/liamvdv/sharedHome/backend"
)

// need some channel to say that DownloadFailed.

func Download(ctx context.Context, service backend.FileCreator, f backend.RemoteFile) {
	
}

// func DownloadFile(ctx context.Context, enc *stream.Encryptor, remote backend.FileReader, h *fs.FileHeader, relpath string) {
// 	hashFp := enc.HashFilepath(relpath)
// 	rh := backend.RemoteFileHeader{
// 		Relpath: hashFp,
// 		Name: filepath.Base(hashFp),
// 	}
// 	dlpath := filepath.Join(config.Home, relpath[:len(relpath)-len(h.Name)], "~" + h.Name)
// 	f, err := os.OpenFile(dlpath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
// 	if err != nil {
// 		if errs.Is(err, os.ErrExist) {
// 			// cannot know if user uses tht prefix, thus cannot just overwrite.
// 			log.Printf("Cannot download file %s to %q, because it would overwrite existing file %q.", h.Name, dlpath)
// 			return
// 		}
// 		log.Printf("Unexpected error occured while opening file %q: %v\n", dlpath, err)
// 		// state.FailedDownload(h) TODO(liamvdv): log this to state file. User should be informed
// 		return
// 	}
// 	var success bool
// 	defer func(fp string) {
// 		if !success {
// 			if err := os.Remove(dlpath); err != nil {
// 				log.Printf("Cannot remove failed download fragment %q: %v\n", dlpath, err)
// 			}
// 			return
// 		}
// 	}(dlpath)

// 	// TODO(liamvdv): Also wrap with compression.

// 	dst, err := stream.NewStreamDecryptor(f, stream.DynamicStreamHeader, enc.Key())

// 	err = retry.Do(3, retry.ExponentialBackoff, func() error {
// 		if err := f.Seek(0, io.SeekStart); err != nil {
// 			return retry.Stop
// 		}
// 		return remote.ReadFile(ctx, h, dst)
// 	})

// 	if err != nil {
// 		log.Println("unable to download ...")
// 		return
// 	}
// 	// ....
// 	if err := os.Rename(dlpath, h.abspath); err != nil {
// 		log.Println("Failed to replace file with complete download fragment.")
// 		return
// 	}
// 	success = true
// 	return

// }
