package native

/*
#cgo CFLAGS: -I${SRCDIR}/zlib/
#cgo windows LDFLAGS: ${SRCDIR}/libs/winlibz.a
#cgo linux LDFLAGS: ${SRCDIR}/libs/linuxlibz.a
#cgo darwin LDFLAGS: ${SRCDIR}/libs/darwinlibz.a

#include "zlib.h"
#include <stdlib.h>
#include <stdint.h>

typedef unsigned char b;

z_stream* newStream() {
	return (z_stream*) calloc(1, sizeof(z_stream));
}

void freeMem(z_stream* s) {
	free(s);
}

int64_t getProcessed(z_stream* s, int64_t inSize) {
	return inSize - s->avail_in;
}

int64_t getCompressed(z_stream* s, int64_t outSize) {
	return outSize - s->avail_out;
}

void prepare(z_stream* s,  int64_t inPtr, int64_t inSize, int64_t outPtr, int64_t outSize) {
    s->avail_in = inSize;
    s->next_in = (b*) inPtr;

    s->avail_out = outSize;
    s->next_out = (b*) outPtr;
}
*/
import "C"
import (
	"unsafe"
)

type processor struct {
	s            *C.z_stream
	hasCompleted bool
	readable     int
	isClosed     bool
}

func newProcessor() processor {
	return processor{C.newStream(), false, 0, false}
}

func (p *processor) prepare(inPtr uintptr, inSize int, outPtr uintptr, outSize int) {
	C.prepare(
		p.s,
		toInt64(int64(inPtr)),
		intToInt64(inSize),
		toInt64(int64(outPtr)),
		intToInt64(outSize),
	)
}

func (p *processor) close() {
	C.freeMem(p.s)
	p.s = nil
	p.isClosed = true
}

func (p *processor) process(in []byte, buf []byte, condition func() bool, zlibProcess func() C.int, specificReset func() C.int) (int, []byte, error) {
	inMem := &in[0]
	inIdx := 0
	p.readable = len(in) - inIdx

	outIdx := 0

	for condition() {
		buf = grow(buf, minWritable)

		outMem := startMemAddress(buf)

		readMem := uintptr(unsafe.Pointer(inMem)) + uintptr(inIdx)
		readLen := len(in) - inIdx
		p.readable = readLen
		writeMem := uintptr(unsafe.Pointer(outMem)) + uintptr(outIdx)
		writeLen := cap(buf) - outIdx

		p.prepare(readMem, readLen, writeMem, writeLen)

		ok := zlibProcess()
		switch ok {
		case C.Z_STREAM_END:
			p.hasCompleted = true
		case C.Z_OK:
		default:
			return inIdx, buf, determineError(errProcess, ok)
		}

		inIdx += int(C.getProcessed(p.s, intToInt64(readLen)))
		outIdx += int(C.getCompressed(p.s, intToInt64(writeLen)))
		buf = buf[:outIdx]
	}

	p.hasCompleted = false

	if ok := specificReset(); ok != C.Z_OK {
		return inIdx, buf, determineError(errReset, ok)
	}

	return inIdx, buf, nil
}
