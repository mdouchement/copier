package copier

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// Bps Bytes per second
	Bps int = 1 << (10 * (iota))
	// KBps Kilo Bytes per second
	KBps
	// MBps Mega Bytes per second
	MBps
	// GBps Giga Bytes per second
	GBps
)

const (
	separator = string(filepath.Separator)
	// StatusAlreadyExist - the destination file exist and is more recent.
	StatusAlreadyExist = "already exist"
	// StatusCopied - file is copied.
	StatusCopied = "copied"
	// StatusOverwritten - destination file is overwritten due to older timpstamp or different size.
	StatusOverwritten = "overwritten"
	// StatusFailed - copy has failed.
	StatusFailed = "failed"
)

// GetBaseDirectory return the common path of the given paths.
func GetBaseDirectory(paths []string) (string, error) {
	if len(paths) < 2 {
		return "", fmt.Errorf("GetBaseDirectory: paths's length must be greater than 2")
	}

	base := paths[0]
	for _, filename := range paths[1:] {
		base = GetBasePath(base, filename)
	}

	return base, nil
}

// GetBasePath returns the identical part of the given two paths.
func GetBasePath(input, output string) string {
	index := 0
	in := strings.Split(input, separator)
	out := strings.Split(output, separator)

	for i, e := range in {
		if i >= len(out) || out[i] != e {
			break
		} else {
			index++
		}
	}

	if separator == "/" {
		// Tricks for joining with root
		in[0] = "/"
	}
	if separator == "\\" {
		// Tricks for joining with drive letter (windows)
		in[0] = fmt.Sprintf("%s\\", in[0])
	}

	return filepath.Join(in[0:index]...)
}

// MkdirAllWithFilename creates all parent directories of the given filepath
func MkdirAllWithFilename(path string) {
	MkdirAll(filepath.Dir(path))
}

// MkdirAll creates the given directory and its parrents
func MkdirAll(path string) {
	if !Exists(path) {
		os.MkdirAll(path, 0755)
	}
}

// Exists checks whether the path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true // ignoring error
}

// Size return the size of the given path.
func Size(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// MoreRecent checks if the dst file is more recent and have the same size of the src file.
func MoreRecent(dst, src string) (bool, error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return false, fmt.Errorf("non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !(dfi.Mode().IsRegular()) {
		return false, fmt.Errorf("non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
	}

	return dfi.Size() == sfi.Size() && dfi.ModTime().After(sfi.ModTime()), nil
}

//----------------//
// ProxyReader    //
//----------------//

// A ProxyReader is proxified reader.
type ProxyReader struct {
	m      sync.Mutex
	reader io.Reader
	writer io.Writer
	closed bool
	c      chan int
}

// NewProxyReader new proxyed reader.
func NewProxyReader(r io.Reader) *ProxyReader {
	return &ProxyReader{
		reader: r,
		c:      make(chan int, 10000),
	}
}

// Read implements io.Reader
func (r *ProxyReader) Read(b []byte) (n int, err error) {
	n, err = r.reader.Read(b)

	if !r.Closed() {
		r.c <- n
	}
	return
}

// Close the reader when it implements io.Closer
func (r *ProxyReader) Close() (err error) {
	if r.Closed() {
		return nil
	}

	if r.reader != nil {
		if closer, ok := r.reader.(io.Closer); ok {
			return closer.Close()
		}
	}
	close(r.c)

	r.m.Lock()
	defer r.m.Unlock()
	r.closed = true
	return
}

// Closed returns the reader status.
func (r *ProxyReader) Closed() bool {
	r.m.Lock()
	defer r.m.Unlock()
	return r.closed
}

// ReadChunk returns the number of read bytes for a chunk.
func (r *ProxyReader) ReadChunk() <-chan int {
	return r.c
}
