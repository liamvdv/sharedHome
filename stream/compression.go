package stream

import (
	"compress/gzip"
	"io"
)

// Push Compression support behind, do later!!!

/*

	It might be better to use an compression alogrithm before pushing the data
	into the encryptor. Thus we reduce the stress on the network and achive faster
	transmittion times. On the other had, it will strain the cpu when pushing to the server.

	Assume more reads than writes of a remote file. Thus compression might take longer
	but decompression should be fast.

*/

// store original filename and other meta data in compression header.

type Compression struct {
	Source           io.Reader
	writer           *gzip.Writer
	CompressedSource io.Reader
}
type Decompression struct {
	Source io.Reader
	Writer io.Writer
}

func NewStreamDecompression(src io.Reader) (*Decompression, error) {
	return nil, nil
}

func NewCompression(src io.Reader) (*Compression, error) {
	pr, pw := io.Pipe()
	gw := gzip.NewWriter(pw)
	return &Compression{
		Source:           src,
		writer:           gw,
		CompressedSource: pr,
	}, nil
}

func (c *Compression) Read(buf []byte) (int, error) {
	n, rErr := c.Source.Read(buf)
	if n > 0 {
		c.writer.Write(buf[:n])
		c.writer.Flush()
		return c.CompressedSource.Read(buf)
	}
	return 0, rErr
}

// // or use pipe?
// AnyIoReader, err := os.Open("")

// src := AnyIoReader
// pReader, pWriter := io.Pipe()
// gw := gzip.Writer(pWriter)
// go func() {
// 	n, err := io.Copy(gw, src)
// 	gw.Close()
// 	pWriter.Close()
// }
// // now use
// pReader
