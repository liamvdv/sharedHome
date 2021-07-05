package stream

import (
	"bytes"
	"compress/gzip"
	"io"
)

/*

	It might be better to use an compression alogrithm before pushing the data
	into the encryptor. Thus we reduce the stress on the network and achive faster
	transmittion times. On the other had, it will strain the cpu when pushing to the server.

	Assume more reads than writes of a remote file. Thus compression might take longer
	but decompression should be fast.

*/

type StreamCompression struct {
	Source io.Reader
	Writer io.Writer
}

type StreamDecompression struct {
	Source io.Reader
	Writer io.Writer
}

func NewStreamDecompression(src io.Reader) (*StreamDecompression, error) {
	return nil, nil
}

func NewStreamCompression(src io.Reader) (*StreamCompression, error) {
	var buf bytes.Buffer
	return &StreamCompression{
		Source: src,
		Writer: gzip.NewWriter(&buf), 
	}, nil
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