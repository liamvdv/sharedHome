package stream

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
)

const (
	// Block size must be of length IV_SIZE.
	IV_SIZE int = 16 // bytes
	// Buffer must be multiple of block size.
	BUFFER_SIZE = 4096 // bytes
)

// Versioning the files ensures that the encryption is never broken,
// even if this software switches gears.
const VERSION_SIZE = 2 // bytes
var (
	// Version = 0 is a placeholder.

	// VERSION_1 is the first version in the versioning model.
	VERSION_1 = [VERSION_SIZE]byte{0x00, 0x01}
)

const headerSize = VERSION_SIZE + IV_SIZE

func NewEncryption(src io.Reader, key []byte) (*Encryption, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	return &Encryption{
		Version: VERSION_1,
		Source:  src,
		Block:   block,
		Stream:  cipher.NewCTR(block, iv),
		Mac:     hmac.New(sha256.New, key),
		Iv:      iv,
	}, nil
}

// Encryption is an io.Reader wrapping an io.Reader.
type Encryption struct {
	Version [VERSION_SIZE]byte
	Source  io.Reader
	Block   cipher.Block
	Stream  cipher.Stream
	Mac     hash.Hash
	Iv      []byte
}

func (enc *Encryption) Read(buf []byte) (int, error) {
	n, rErr := enc.Source.Read(buf)
	if n > 0 {
		enc.Stream.XORKeyStream(buf[:n], buf[:n])

		m, err := enc.Mac.Write(buf[:n])
		if err != nil {
			return 0, fmt.Errorf("cannot write to mac: %w", err)
		}
		if m != n {
			return 0, fmt.Errorf("cannot write all bytes to hmac")
		}

		return n, rErr
	}
	return 0, io.EOF
}

func (enc *Encryption) Header() EncryptionHeader {
	return EncryptionHeader{
		Version: enc.Version,
		Iv:      enc.Iv,
	}
}

type EncryptionHeader struct {
	// Version is a 2 byte separation tool.
	Version [VERSION_SIZE]byte
	// Iv is a 4 byte random initialisation vector required by AES
	Iv []byte
}

// WriteTo always returns headerSize and nil.
func (h EncryptionHeader) WriteTo(w io.Writer) (int64, error) {
	w.Write(h.Version[:])
	w.Write(h.Iv)
	return int64(headerSize), nil
}

// Footer must only be called after the content of the stream has been fully read.
func (enc *Encryption) Footer() EncryptionFooter {
	return EncryptionFooter{
		Mac: enc.Mac.Sum(nil),
	}
}

type EncryptionFooter struct {
	Mac []byte
}

func (f EncryptionFooter) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(f.Mac)
	return int64(n), err
}

/*========================================== Decryption ==========================================*/

func ReadHeader(src io.Reader) (*EncryptionHeader, error) {
	buf := make([]byte, headerSize)
	n, rErr := src.Read(buf)
	if n != headerSize {
		return nil, fmt.Errorf("cannot read full header.")
	}
	h := EncryptionHeader{Iv: buf[VERSION_SIZE:]}
	copy(h.Version[:], buf[:VERSION_SIZE])
	return &h, rErr
}

func NewDecryption(key []byte, src io.Reader, h EncryptionHeader) (*Decryption, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &Decryption{
		Version: h.Version,
		Source:  src,
		Block:   block,
		Stream:  cipher.NewCTR(block, h.Iv),
		Mac:     hmac.New(sha256.New, key),
	}, nil
}

type Decryption struct {
	Version [VERSION_SIZE]byte
	Source  io.Reader
	Block   cipher.Block
	Stream  cipher.Stream
	Mac     hash.Hash
}

func (dec *Decryption) Read(buf []byte) (int, error) {
	n, rErr := dec.Source.Read(buf)
	if n > 0 {
		m, err := dec.Mac.Write(buf[:n])
		if err != nil {
			return 0, fmt.Errorf("cannot write to mac: %w", err)
		}
		if m != n {
			return 0, fmt.Errorf("cannot write all bytes to hmac")
		}

		dec.Stream.XORKeyStream(buf[:n], buf[:n])
		return n, rErr
	}
	return 0, io.EOF
}

// ValidMac must only be called after all encoded content has been read.
func (dec *Decryption) ValidMac(mac []byte) bool {
	return bytes.Compare(mac, dec.Mac.Sum(nil)) == 0
}
