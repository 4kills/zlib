package zlib

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

// real world data benchmarks

const decompressedMcPacketsLoc = "test/mc_packets/decompressed_mc_packets.json"

var decompressedMcPackets [][]byte

func BenchmarkWriteBytesAllMcPacketsDefault(b *testing.B) {
	loadPacketsIfNil(&decompressedMcPackets, decompressedMcPacketsLoc)

	benchmarkWriteBytesMcPacketsGeneric(decompressedMcPackets, b)
}

func benchmarkWriteBytesMcPacketsGeneric(input [][]byte, b *testing.B) {
	w := NewWriter(nil)
	defer w.Close()

	reportBytesPerChunk(input, b)

	for i := 0; i < b.N; i++ {
		for _, v := range input {
			w.WriteBytes(v)
		}
	}
}

func BenchmarkWriteAllMcPacketsDefault(b *testing.B) {
	loadPacketsIfNil(&decompressedMcPackets, decompressedMcPacketsLoc)
	w := NewWriter(&bytes.Buffer{})

	benchmarkWriteMcPacketsGeneric(w, decompressedMcPackets, b)
}

func BenchmarkWriteAllMcPacketsDefaultStd(b *testing.B) {
	loadPacketsIfNil(&decompressedMcPackets, decompressedMcPacketsLoc)
	w := zlib.NewWriter(&bytes.Buffer{})

	benchmarkWriteMcPacketsGeneric(w, decompressedMcPackets, b)
}

func benchmarkWriteMcPacketsGeneric(w TestWriter, input [][]byte, b *testing.B) {
	reportBytesPerChunk(input, b)

	defer w.Close()

	for i := 0; i < b.N; i++ {
		for _, v := range input {
			w.Write(v)
		}
	}
}

// laboratory condition benchmarks

func BenchmarkWriteBytes64BBestCompression(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(64), BestCompression, b)
}

func BenchmarkWriteBytes8192BBestCompression(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(8192), BestCompression, b)
}

func BenchmarkWriteBytes65536BBestCompression(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(65536), BestCompression, b)
}

func BenchmarkWriteBytes64BBestSpeed(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(64), BestSpeed, b)
}

func BenchmarkWriteBytes8192BBestSpeed(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(8192), BestSpeed, b)
}

func BenchmarkWriteBytes65536BBestSpeed(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(65536), BestSpeed, b)
}

func BenchmarkWriteBytes64BDefault(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(64), DefaultCompression, b)
}

func BenchmarkWriteBytes8192BDefault(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(8192), DefaultCompression, b)
}

func BenchmarkWriteBytes65536BDefault(b *testing.B) {
	benchmarkWriteBytesLevel(xByte(65536), DefaultCompression, b)
}

func benchmarkWriteBytesLevel(input []byte, level int, b *testing.B) {
	w, _ := NewWriterLevel(nil, level)
	defer w.Close()

	for i := 0; i < b.N; i++ {
		w.WriteBytes(input)

		b.StopTimer()
		w.Reset(nil) // to ensure there are no caching effects for the same packet over and over again
		b.StartTimer()
	}
}

func BenchmarkWrite64BBestCompression(b *testing.B) {
	benchmarkWriteLevel(xByte(64), BestCompression, b)
}

func BenchmarkWrite8192BBestCompression(b *testing.B) {
	benchmarkWriteLevel(xByte(8192), BestCompression, b)
}

func BenchmarkWrite65536BBestCompression(b *testing.B) {
	benchmarkWriteLevel(xByte(65536), BestCompression, b)
}

func BenchmarkWrite64BBestSpeed(b *testing.B) {
	benchmarkWriteLevel(xByte(64), BestSpeed, b)
}

func BenchmarkWrite8192BBestSpeed(b *testing.B) {
	benchmarkWriteLevel(xByte(8192), BestSpeed, b)
}

func BenchmarkWrite65536BBestSpeed(b *testing.B) {
	benchmarkWriteLevel(xByte(65536), BestSpeed, b)
}

func BenchmarkWrite64BDefault(b *testing.B) {
	benchmarkWriteLevel(xByte(64), DefaultCompression, b)
}

func BenchmarkWrite8192BDefault(b *testing.B) {
	benchmarkWriteLevel(xByte(8192), DefaultCompression, b)
}

func BenchmarkWrite65536BDefault(b *testing.B) {
	benchmarkWriteLevel(xByte(65536), DefaultCompression, b)
}

func benchmarkWriteLevel(input []byte, level int, b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, len(input)))
	w, _ := NewWriterLevel(buf, level)

	benchmarkWriteLevelGeneric(w, buf, input, b)
}

func BenchmarkWrite64BBestCompressionStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(64), BestCompression, b)
}

func BenchmarkWrite8192BBestCompressionStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(8192), BestCompression, b)
}

func BenchmarkWrite65536BBestCompressionStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(65536), BestCompression, b)
}

func BenchmarkWrite64BBestSpeedStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(64), BestSpeed, b)
}

func BenchmarkWrite8192BBestSpeedStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(8192), BestSpeed, b)
}

func BenchmarkWrite65536BBestSpeedStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(65536), BestSpeed, b)
}

func BenchmarkWrite64BDefaultStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(64), DefaultCompression, b)
}

func BenchmarkWrite8192BDefaultStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(8192), DefaultCompression, b)
}

func BenchmarkWrite65536BDefaultStd(b *testing.B) {
	benchmarkWriteLevelStd(xByte(65536), DefaultCompression, b)
}

// std library zlib in comparison
func benchmarkWriteLevelStd(input []byte, level int, b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, len(input)))
	w, _ := NewWriterLevel(buf, level)

	benchmarkWriteLevelGeneric(w, buf, input, b)
}

func benchmarkWriteLevelGeneric(w TestWriter, buf *bytes.Buffer, input []byte, b *testing.B) {
	defer w.Close()

	for i := 0; i < b.N; i++ {
		w.Write(input)

		b.StopTimer()
		w.Reset(buf) // to ensure there are no caching effects for the same packet over and over again
		buf.Reset()
		b.StartTimer()
	}
}

// HELPER FUNCTIONS

func reportBytesPerChunk(input [][]byte, b *testing.B) {
	numOfBytes := 0
	for _, v := range input {
		numOfBytes += len(v)
	}
	b.ReportMetric(float64(numOfBytes), "bytes/chunk")
}

func loadPacketsIfNil(packets *[][]byte, loc string) {
	if *packets != nil {
		return
	}
	*packets = loadPackets(loc)
}

type TestWriter interface {
	Write(p []byte) (int, error)
	Close() error
	Flush() error
	Reset(w io.Writer)
}

func loadPackets(loc string) [][]byte {
	var b [][]byte
	jsonFile, _ := os.Open(loc)
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &b)
	return b
}

func xByte(multOf16 int) []byte {
	_16byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf}

	if multOf16 == 0 || multOf16%16 != 0 {
		panic(errors.New("multOf16 is not a valid multiple of 16"))
	}

	xByte := make([]byte, multOf16)
	for i := 0; i < multOf16; i += 16 {
		copy(xByte[i:i+16], _16byte)
	}

	return xByte
}