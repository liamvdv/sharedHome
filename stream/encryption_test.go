package stream_test

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"

	"github.com/liamvdv/sharedHome/stream"
)

func TestEnAndDecodingFile(t *testing.T) {
	inPath := "./testdata/encryptMe.txt"
	inFile, err := os.Open(inPath)
	if err != nil {
		t.Error(err)
	}
	defer inFile.Close() // omit already closed error.

	key := stream.HashKey("super-secret secret-key")
	enc, err := stream.NewEncryption(inFile, []byte(key))
	if err != nil {
		t.Error(err)
	}

	encPath := "./testresult/" + stream.HashName("encryptMe.txt")
	encFile, err := os.Create(encPath)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		_ = encFile.Close()
		err := os.Remove(encPath)
		if err != nil {
			log.Printf("failed to remove temporary file %q: %v\n", encPath, err)
		}
	}()

	bWroteHeader, err := enc.Header().WriteTo(encFile)
	if err != nil {
		t.Error(err)
	}
	bWroteEnc, err := io.Copy(encFile, enc)
	if err != nil {
		t.Error(err)
	}
	// Use
	// if _, err := enc.Footer().WriteTo(dst); err != nil { ... }
	// in production.
	footer := enc.Footer()
	originalHmac := footer.Mac // for comparsion
	if _, err := footer.WriteTo(encFile); err != nil {
		t.Error(err)
	}

	/* ################### Decryption ################### */

	// Find, read and trim hmac.
	// TODO(liamvdv): Validate that hmac is alwayas 32 bytes (256 bits) long if sha256 is used.
	if _, err := encFile.Seek(-32, io.SeekEnd); err != nil {
		t.Error(err)
	}
	receivedHmac := make([]byte, 32)
	n, err := encFile.Read(receivedHmac)
	if err != nil {
		t.Error(err)
		return
	}
	if n != 32 {
		t.Error("Did not read full hmac.")
	}
	if err := encFile.Truncate(bWroteHeader + bWroteEnc); err != nil {
		t.Error(err)
	}

	// Read and decrypt file.
	if _, err := encFile.Seek(0, io.SeekStart); err != nil {
		t.Error(err)
	}
	outPath := "./testresult/encryptMe.txt"
	outFile, err := os.OpenFile(outPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0600) // in production use os.O_Exclude
	if err != nil {
		t.Error(err)
	}
	defer func() {
		_ = outFile.Close()
		err := os.Remove(outPath)
		if err != nil {
			log.Printf("failed to remove temporary file %q: %v\n", outPath, err)
		}
	}()

	h, err := stream.ReadHeader(encFile)
	if err != nil {
		t.Error(err)
	}

	dec, err := stream.NewDecryption(key, encFile, *h)
	if err != nil {
		t.Error(err)
	}

	bWroteDec, err := io.Copy(outFile, dec)
	if err != nil {
		t.Error(err)
	}

	if bWroteEnc != bWroteDec {
		t.Errorf("encrypted %d bytes, only decrypted %d bytes", bWroteDec, bWroteDec)
	}

	if !dec.ValidMac(receivedHmac) {
		t.Errorf(`
hmac not matching
really want: %x
calc   want: %x
calc    got: %x`, originalHmac, receivedHmac, dec.Mac.Sum(nil))
	}

	outFile.Close()
	encFile.Close()

	n, err = compareFileContent(inPath, outPath)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("unencrypted is not equal to en- and decrypted file.")
	}
}

// compareFileContent reads to files into memory and then compares the byte sequence.
func compareFileContent(apath string, bpath string) (int, error) {
	araw, err := os.ReadFile(apath)
	if err != nil {
		return 0, err
	}

	braw, err := os.ReadFile(bpath)
	if err != nil {
		return 0, err
	}
	return bytes.Compare(araw, braw), nil
}

// TODO(liamvdv): Move to integration test with metadata manipulation.
func compareFileMetadata(apath string, bpath string) (bool, error) {
	afi, err := os.Stat(apath)
	if err != nil {
		return false, err
	}
	bfi, err := os.Stat(bpath)
	if err != nil {
		return false, err
	}

	m := afi.Mode() == bfi.Mode()
	n := afi.Name() == bfi.Name()
	s := afi.Size() == bfi.Size()
	mt := afi.ModTime().Equal(bfi.ModTime())
	return m && n && s && mt, nil
}
