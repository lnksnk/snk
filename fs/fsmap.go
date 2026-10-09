package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
)

func FSMap(fsroot, fslocalroot string) (err error) {
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "openFSFile", Path: "", Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}
	if fslocalroot != "" && fslocalroot[len(fslocalroot)-1] != '/' {
		fslocalroot += "/"
	}
	GlobalFSLocalRoot.Store(fsroot, fslocalroot)
	if fslocalroot == "" {
		GlobalFSOpen.Store(fsroot, OpenFileSystemFunc(func(name string) (File, error) {
			if name != "" && len(name) >= len(fsroot) && name[:len(fsroot)] == fsroot {
				if memf, loaded := GlobalMemFile.Load(name); loaded {
					return memf.File(), nil
				}
			}
			return nil, &fs.PathError{Op: "memFileSystem", Path: name, Err: fmt.Errorf("invalid path")}
		}))

		GlobalFSStat.Store(fsroot, StatFileSystemFunc(func(name string) (FileInfo, error) {
			if name != "" && len(name) >= len(fsroot) && name[:len(fsroot)] == fsroot {
				if memf, loaded := GlobalMemFile.Load(name); loaded {
					return memf.Stat()
				}
			}
			return nil, &fs.PathError{Op: "memFileSystem", Path: name, Err: fmt.Errorf("invalid path")}
		}))

		GlobalFSSet.Store(fsroot, SetFileSystemFunc(func(path string, a ...any) (err error) {
			if path != "" && fsroot != "" && fsroot[len(fsroot)-1] == '/' && len(path) >= len(fsroot) && path[:len(fsroot)] == fsroot {
				if ext := filepath.Ext(path); ext != "" {

					var mfrf *memfile
					var mffound MemFile
					if mffound, _ = GlobalMemFile.Load(path); mffound != nil {
						mfrf, _ = mffound.(*memfile)
						if err = mfrf.Set(path, fsroot, a...); err != nil {
							GlobalMemFile.Delete(path)
							return
						}
					}

					if mfrf == nil {
						mfrf = &memfile{lck: &sync.RWMutex{}}
					}

					if err = mfrf.Set(path, fsroot, a...); err == nil {
						if mffound == nil {
							GlobalMemFile.Store(path, mfrf)
						}
						return
					}
					if mffound == nil {
						mfrf.Close()
					}
					return
				}
			}
			return &fs.PathError{Op: "memFileSystem", Path: path, Err: fmt.Errorf("invalid path")}
		}))

		GlobalFSAppend.Store(fsroot, AppendFileSystemFunc(func(path string, a ...any) (err error) {
			if path != "" && fsroot != "" && fsroot[len(fsroot)-1] == '/' && len(path) >= len(fsroot) && path[:len(fsroot)] == fsroot {
				if ext := filepath.Ext(path); ext != "" {
					var mfrf *memfile
					var mffound MemFile
					if mffound, _ = GlobalMemFile.Load(path); mffound != nil {
						mfrf, _ = mffound.(*memfile)
						if err = mfrf.Append(path, fsroot, a...); err != nil {
							GlobalMemFile.Delete(path)
							return
						}
					}

					if mfrf == nil {
						mfrf = &memfile{lck: &sync.RWMutex{}}
					}

					if err = mfrf.Append(path, fsroot, a...); err == nil {
						if mffound == nil {
							GlobalMemFile.Store(path, mfrf)
						}
						return
					}
					if mffound == nil {
						mfrf.Close()
					}
					return
				}
			}
			return &fs.PathError{Op: "memFileSystem", Path: path, Err: fmt.Errorf("invalid path")}
		}))
	}
	return
}
