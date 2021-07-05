/*
AUTHORs Note:
encryption.go and it's content is mainly based on github.com/blend/go-sdk/crypto/stream.go,
which is governed by the MIT licencse and the github.com/blend/go-sdk/LICENSE file.
Attribution to the authors as of 2021/06/22, namely:
Name 				Profile
Will Charczuk		github.com/wcharzuk
Micheal Turner		github.com/mat285
*/

package stream

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	errs "errors"
	"hash"
	"io"
	"os"
	"strings"
)

const (
	path_sep = string(os.PathSeparator)
)

/*
myfile.txt -> sha1("myfile.txt")
..content.. ->
	4 byte Version
	4 byte IV_SIZE
	n byte contents
	32 byte hmac
*/

const (
	// Buffer must be a multiple of block size, which must be equal in length to IV_SIZE
	BUFFER_SIZE      = 4096
	IV_SIZE      int = 16
	VERSION_SIZE     = 4
	header_size      = VERSION_SIZE + IV_SIZE
)

var (
	// every version must have a unique 4 byte sequence.
	VERSION_0           = []byte{0, 0, 0, 7}
	VERSION_PLACEHOLDER = []byte("fake")
)

type Encryptor struct {
	tmpDirpath string
	key        []byte
}

func (enc *Encryptor) Key() []byte {
	return enc.key
}

func NewEncrytor(plainkey string) *Encryptor {
	k := sha256.Sum256([]byte(plainkey))
	return &Encryptor{
		tmpDirpath: "", // TODO(liamvdv): tmpDirpath?
		key:        k[:],
	}
}

func (E *Encryptor) HashFilename(name string) string {
	sha := sha1.Sum([]byte(name))
	return base64.URLEncoding.EncodeToString(sha[:])
}

// EncryptFilepath returns the encrypted filepath, no trailing os.PathSeperator.
func (E *Encryptor) HashFilepath(fp string) string {
	paths := strings.Split(fp, path_sep)
	for idx := range paths[1:] {
		paths[1+idx] = E.HashFilename(paths[1+idx])
	}

	return path_sep + strings.Join(paths, path_sep)
}

func (E *Encryptor) Close() error {
	return os.RemoveAll(E.tmpDirpath)
}

// // Returned file must be closed by consumer. Tmp File is stored und relative fp
// // to Encryptor tmp file directory.
// func (E *Encryptor) Encrypt(fp string) (*os.File, error) {
// 	dp := filepath.Dir(fp)
// 	encDp := E.HashFilepath(dp)

// 	name := filepath.Base(fp)
// 	enc := E.HashFilename(name)

// 	localEncDp := filepath.Join(E.tmpDirpath, encDp)
// 	localEncFp := filepath.Join(localEncDp, enc)

// 	if err := os.MkdirAll(localEncDp, 0700); err != nil {
// 		panic(err)
// 	}

// 	outFile, err := os.OpenFile(localEncFp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
// 	if err != nil {
// 		return nil, err
// 	}

// 	inFile, err := os.Open(fp)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer inFile.Close() //TODO: May also error

// 	encReader, err := NewStreamEncryptor(inFile, E.key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	header := encReader.Header()
// 	outFile.Write(header.Version)
// 	outFile.Write(header.Iv)

// 	if _, err := io.Copy(outFile, encReader); err != nil && err != io.EOF {
// 		return nil, err
// 	}

// 	// Rewind file pointer to start to read from it.
// 	if _, err := outFile.Seek(0, io.SeekStart); err != nil {
// 		return nil, err
// 	}

// 	return outFile, nil
// }

// func (E *Encryptor) Decrypt(src io.Reader) (*StreamDecryption, error) {
// 	header, err := ReadHeader(src)
// 	if err != nil {
// 		return nil, err
// 	}
// 	decReader, err := NewStreamDecryptor(src, header, E.key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return decReader, nil
// }

/*
Usage - encryption
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	var success bool
	defer func() {
		if !success {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}
	}
	src, err := stream.NewStreamEncryption(f, e.KeyHash)
	if err != nil {
		return nil, err
	}

	dst, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}

	// write content
	_, _ = src.Header().WriteTo(dst) // header
	_, err := io.Copy(dst, src) // content
	if err != nil {
		return nil, err
	}
	_, _ = dst.Write(src.Mac.Sum(nil)) // hmac


	if err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	success = true

	remote.Upload(..., dst)

	return f, nil
}

Usage - decryption
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	defer SaveClose(f)

	// download file => direct mapping, no internal buffering
	r, w := io.Pipe()
	go remote.Download(w)

	h, err := ReadHeader(r)
	src, err := stream.NewStreamDecryption(r, h, e.KeyHash)
	if err != nil {
		return nil, err
	}
	// need to drop the hmac

	_, err := io.Copy(f, src)
*/

// TODO: Try to read d


type EncryptionStreamHeader struct {
	Version []byte // 4 byte
	Iv      []byte // 4 byte
}

// WriteTo does always return header_size and nil
func (h EncryptionStreamHeader) WriteTo(w io.Writer) (int64, error) {
	w.Write(h.Version)
	w.Write(h.Iv)
	return int64(header_size), nil
}

func NewStreamEncryption(src io.Reader, key []byte) (*StreamEncryption, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	mac := hmac.New(sha256.New, key)

	return &StreamEncryption{
		Version: VERSION_0,
		Source:  src,
		Block:   block,
		Stream:  stream,
		Mac:     mac,
		Iv:      iv,
	}, nil
}

type StreamEncryption struct {
	Version []byte
	Source  io.Reader
	Block   cipher.Block
	Stream  cipher.Stream
	Mac     hash.Hash
	Iv      []byte
}

func (s *StreamEncryption) Read(buf []byte) (int, error) {
	n, rErr := s.Source.Read(buf)
	if n > 0 {
		s.Stream.XORKeyStream(buf[:n], buf[:n])
		err := writeHash(s.Mac, buf[:n])
		if err != nil {
			return n, errs.New(err.Error())
		}
		return n, rErr
	}
	return 0, io.EOF
}

// Header returns the version and the iv, if they were already read.
func (s *StreamEncryption) Header() EncryptionStreamHeader {
	return EncryptionStreamHeader{
		Version: s.Version,
		Iv:      s.Iv,
	}
}

// DynamicStreamHeader will make the decryption struct detect the header itself
// on reads from it.
var DynamicDecryptHeader = &EncryptionStreamHeader{
	Version: VERSION_PLACEHOLDER,
}

func NewStreamDecryption(src io.Reader, header *EncryptionStreamHeader, key []byte) (*StreamDecryption, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(header.Version, VERSION_PLACEHOLDER) == 0 {
		// detect version dynamically on read calls, so that Header() doesn't have to be called before NewStreamDecryptor.
		return &StreamDecryption{
			Version: VERSION_PLACEHOLDER,
			Source:  src,
			Block:   block,
		}, nil
	}
	stream := cipher.NewCTR(block, header.Iv)
	mac := hmac.New(sha256.New, key)

	return &StreamDecryption{
		Version: header.Version,
		Source:  src,
		Block:   block,
		Stream:  stream,
		Mac:     mac,
	}, nil
}

type StreamDecryption struct {
	Version []byte
	Source  io.Reader
	Block   cipher.Block
	Stream  cipher.Stream
	Mac     hash.Hash
}

func (s *StreamDecryption) Read(buf []byte) (int, error) {
	n, rErr := s.Source.Read(buf)

	// dynamically detect header. Must work on first read.
	// Empty files will be treaded as errors, should at least have a header.
	if bytes.Compare(s.Version, VERSION_PLACEHOLDER) == 0 {
		err := dynamicReadInitHeader(buf[:n], s)
		if err != nil {
			return -1, err
		}
		if n > header_size {
			n2 := n - header_size
			s.Stream.XORKeyStream(buf[:n2], buf[header_size:n])
			return n2, rErr
		}
	}

	if n > 0 {
		s.Stream.XORKeyStream(buf[:n], buf[:n])
		return n, rErr
	}
	return 0, io.EOF
}

// reads and inits dynamic header.
func dynamicReadInitHeader(buf []byte, s *StreamDecryption) error {
	if len(buf) >= header_size {
		h, err := ReadHeader(bytes.NewReader(buf[:header_size]))
		if err != nil {
			return err
		}
		s.Version = h.Version
		s.Stream = cipher.NewCTR(s.Block, h.Iv)
		return nil
	}
	return errs.New("Can not read full header.") // buffer to small or not enough to read.
}

func ReadHeader(src io.Reader) (*EncryptionStreamHeader, error) {
	bHeader := make([]byte, header_size)
	n, rErr := src.Read(bHeader)
	if n != header_size {
		return nil, errs.New("Can not read full header.")
	}

	return &EncryptionStreamHeader{
		Version: bHeader[:IV_SIZE],
		Iv:      bHeader[VERSION_SIZE:],
	}, rErr
}

func writeHash(mac hash.Hash, p []byte) error {
	m, err := mac.Write(p)
	if err != nil {
		return errs.New(err.Error())
	}
	if m != len(p) {
		return errs.New("could not write all bytes to hmac")
	}
	return nil
}
