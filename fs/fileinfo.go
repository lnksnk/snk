package fs

import (
	"io/fs"
	"strings"
	"time"
)

type FileInfo interface {
	Name() string       // base name of the file
	Path() string       // path of th file
	Root() string       // root of the file
	Base() string       // base root of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() fs.FileMode  // file mode bits (may return 0)
	ModTime() time.Time // modification time
	IsDir() bool        // indicate if file info refer to a directory
	Sys() any           // underlying data source (can return nil)
}

type StatFileInfo interface {
	Stat(name string) (FileInfo, error)
}

type StatFileInfoFunc func(name string) (FileInfo, error)

func (stf StatFileInfoFunc) Stat(name string) (FileInfo, error) {
	return stf(name)
}

type fileinfo struct {
	name    string
	path    string
	root    string
	base    string
	szs     int64
	isdir   bool
	modTime time.Time
}

// Base implements [FileInfo].
func (f *fileinfo) Base() string {
	return f.base
}

// IsDir implements [FileInfo].
func (f *fileinfo) IsDir() bool {
	return f.isdir
}

// ModTime implements [FileInfo].
func (f *fileinfo) ModTime() time.Time {
	return f.modTime
}

// Mode implements [FileInfo].
func (f *fileinfo) Mode() fs.FileMode {
	return 0
}

// Name implements [FileInfo].
func (f *fileinfo) Name() string {
	return f.name
}

// Path implements [FileInfo].
func (f *fileinfo) Path() string {
	return f.path
}

// Root implements [FileInfo].
func (f *fileinfo) Root() string {
	return f.root
}

// Size implements [FileInfo].
func (f *fileinfo) Size() int64 {
	return f.szs
}

// Sys implements [FileInfo].
func (f *fileinfo) Sys() any {
	return nil
}

func GenFileInfo(path, base string, szs int64, isdir bool, modTime time.Time) FileInfo {
	var name, root string
	if path != "" {
		if pthi, pthli := strings.Index(path, "/"), strings.LastIndex(path, "/"); pthli > -1 {
			name = path[pthli+1:]
			root = path[:pthli+1]
			base = path[:pthi+1]
		} else {
			name = path
			root = "/"
			path = "/" + path
		}
		goto info
	}

	if path == "" {
		root = "/"
		base = "/"
		name = ""
		path += name
		goto info
	}

info:
	return &fileinfo{name: name, path: path, root: root, base: base, szs: szs, isdir: isdir, modTime: modTime}
}
