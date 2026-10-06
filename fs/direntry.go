package fs

type DirEntry interface {
	// Name returns the name of the file (or subdirectory) described by the entry.
	// This name is only the final element of the path (the base name), not the entire path.
	// For example, Name would return "hello.go" not "home/gopher/hello.go".
	Name() string

	// IsDir reports whether the entry describes a directory.
	IsDir() bool

	// Info returns the FileInfo for the file or subdirectory described by the entry.
	// The returned FileInfo may be from the time of the original directory read
	// or from the time of the call to Info.
	Info() (FileInfo, error)
}

type direntry struct {
	fi FileInfo
}

func (de *direntry) Name() string {
	if fi := de.fi; fi != nil {
		return fi.Name()
	}
	return ""
}

func (de *direntry) IsDir() bool {
	if fi := de.fi; fi != nil {
		return fi.IsDir()
	}
	return false
}

func (de *direntry) Info() (FileInfo, error) {
	return de.fi, nil
}

func dirEntry(fi FileInfo) DirEntry {
	return &direntry{fi: fi}
}
