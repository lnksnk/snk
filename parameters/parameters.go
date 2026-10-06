package parameters

import (
	"bufio"
	"io"
	"mime/multipart"
	http "net/http"
	"net/textproto"
	url "net/url"
	"strings"

	snkio "github.com/lnksnk/snk/io"
)

type File interface {
	Name() string
	Reader() io.Reader
	Size() int64
	Close() error
	Header() textproto.MIMEHeader
}

type file struct {
	name    string
	size    int64
	orgrdr  io.ReadSeekCloser
	flhdr   *multipart.FileHeader
	mimehdr textproto.MIMEHeader
}

// Close implements File.
func (f *file) Close() (err error) {
	if f == nil {
		return
	}
	f.flhdr = nil
	f.mimehdr = nil
	orgrdr := f.orgrdr
	f.orgrdr = nil
	if orgrdr != nil {
		err = orgrdr.Close()
	}
	return
}

// Header implements File.
func (f *file) Header() textproto.MIMEHeader {
	if f == nil {
		return nil
	}
	return f.mimehdr
}

// Size implements File.
func (f *file) Size() int64 {
	if f == nil {
		return 0
	}
	return f.size
}

func (f *file) Name() string {
	if f == nil {
		return ""
	}
	return f.name
}

func (f *file) Reader() io.Reader {
	if f == nil {
		return snkio.ReadFunc(func(p []byte) (n int, err error) {
			return 0, io.EOF
		})
	}
	orgrdr := f.orgrdr
	if orgrdr == nil {
		if flhdr := f.flhdr; flhdr != nil {
			r, rerr := flhdr.Open()
			if rerr != nil {
				return snkio.ReadFunc(func(p []byte) (n int, err error) {
					return 0, rerr
				})
			}
			f.orgrdr = r
			return f.orgrdr
		}
		return snkio.ReadFunc(func(p []byte) (n int, err error) {
			return 0, io.EOF
		})
	}
	return orgrdr
}

func nextfile(flhdr *multipart.FileHeader) File {
	return &file{name: flhdr.Filename, size: flhdr.Size, flhdr: flhdr}
}

type Parameters interface {
	Keys() []string
	FileKeys() []string
	Set(string, bool, ...string)
	Empty(...string) bool
	Exist(string) bool
	Remove(string) []string
	SetFile(string, bool, ...interface{})
	FileExist(string) bool
	FileEmpty(...string) bool
	RemoveFile(string) []File
	Get(string, ...int) []string
	String(string, string, ...int) string
	GetFile(string, ...int) []File
	ClearAll()
	Clear()
	ClearFiles()
	Type(string) string
}

// Parameters -> structure containing parameters
type parameters struct {
	urlkeys        snkio.Map[string, bool]     //*sync.Map
	standard       snkio.Map[string, []string] //map[string][]string
	standardcount  int
	filesdata      snkio.Map[string, []File] //map[string][]interface{}
	filesdatacount int
}

var EmptyParameters = &parameters{}

var emptyParmVal = []string{}
var emptyParamFile = []File{}

// Keys - list of standard parameters names (keys)
func (params *parameters) Keys() (keys []string) {
	if params != nil {
		if standard, standardcount := params.standard, params.standardcount; standard != nil && standardcount > 0 {
			if keys == nil {
				keys = make([]string, standardcount)
			}
			ki := 0
			standard.Range(func(key string, value []string) bool {
				if ki < standardcount {
					keys[ki] = key
					ki++
				}
				return true
			})
		}
	}
	return keys
}

// FileKeys - list of file parameters names (keys)
func (params *parameters) FileKeys() (keys []string) {
	if params != nil {
		if filesdata, filesdatacount := params.filesdata, params.filesdatacount; filesdata != nil && filesdatacount > 0 {
			if keys == nil {
				keys = make([]string, filesdatacount)
			}
			ki := 0
			filesdata.Range(func(key string, value []File) bool {
				if ki < filesdatacount {
					keys[ki] = key
					ki++
				}
				return true
			})
		}
	}
	return keys
}

// Set -> set or append parameter value
// pname : name
// pvalue : value of strings to add
// clear : clear existing value of parameter
func (params *parameters) Set(pname string, clear bool, pvalue ...string) {
	storeParameter(params, false, pname, clear, pvalue...)
}

func storeParameter(params *parameters, isurl bool, pname string, clear bool, pvalue ...string) {
	if pname = strings.ToUpper(strings.TrimSpace(pname)); pname == "" {
		return
	}

	if params != nil {
		standard, urlkeys := params.standard, params.urlkeys
		if urlkeys == nil {
			urlkeys = snkio.NewMap[string, bool]()
			params.urlkeys = urlkeys
		}
		if standard == nil {
			standard = snkio.NewMap[string, []string]() // make(map[string][]string)
			params.standard = standard
		}
		if val, ok := standard.Load(pname); ok {
			if clear {
				standard.Swap(pname, []string{})
				val, _ = standard.Load(pname)
				urlkeys.LoadAndDelete(pname)
			}
			var valsarr = val
			if len(pvalue) > 0 {
				valsarr = append(valsarr, pvalue...)
				urlkeys.Store(pname, isurl)
			}
			params.standard.Swap(pname, valsarr)
		} else {
			if len(pvalue) > 0 {
				urlkeys.Store(pname, isurl)
				params.standard.Store(pname, pvalue[:])
				params.standardcount++
			} else {
				params.standard.Store(pname, []string{})
				params.standardcount++
			}
		}
	}
}

// Exist -> check if parameter exist
// pname : name
func (params *parameters) Exist(pname string) bool {
	if pname = strings.ToUpper(strings.TrimSpace(pname)); pname == "" {
		return false
	}
	standard := params.standard
	if standard == nil {
		return false
	}
	_, ok := standard.Load(pname)
	return ok
}

// Type -> check if parameter was loaded as a url/standard parameter
// pname : name
func (params *parameters) Type(pname string) string {
	if pname = strings.ToUpper(strings.TrimSpace(pname)); pname == "" {
		return ""
	}
	standard, urlkeys := params.standard, params.urlkeys
	if urlkeys == nil && standard == nil {
		return ""
	}
	if isurlv, ok := urlkeys.Load(pname); ok {
		if isurlv {
			return "url"
		}
	}
	if _, ok := standard.Load(pname); ok {
		return "std"
	}
	return ""
}

// Remove  -> remove parameter and return any slice of string value
func (params *parameters) Remove(pname string) (value []string) {
	if pname = strings.ToUpper(strings.TrimSpace(pname)); pname == "" {
		return
	}
	standard, urlkeys := params.standard, params.urlkeys
	if standard == nil {
		return
	}
	if stdval, ok := standard.LoadAndDelete(pname); ok {
		if urlkeys != nil {
			urlkeys.LoadAndDelete(pname)
		}
		params.standardcount--
		value = stdval
	}
	return
}

// Empty  -> return true if there are no parameters
func (params *parameters) Empty(pname ...string) (empty bool) {
	if params != nil {
		empty = true
		if standard := params.standard; standard != nil {
			prmsl := len(pname)
			standard.Range(func(key string, value []string) bool {
				if prmsl == 0 {
					empty = !true
					return empty
				}
				if prmsl == 1 && pname[0] != "" {
					if strings.EqualFold(key, pname[0]) {
						if len(value) == 0 {
							return empty
						}
						for _, v := range value {
							if empty = v == ""; empty {
								continue
							}
							return empty
						}
						return empty
					}
				}
				return empty
			})
		}
		return
	}
	return true
}

// SetFile -> set or append file parameter value
// pname : name
// pfile : value of interface to add either FileHeader from mime/multipart or any io.Reader implementation
// clear : clear existing value of parameter
func (params *parameters) SetFile(pname string, clear bool, pfile ...interface{}) {
	if params != nil {
		if pname = strings.ToUpper(strings.TrimSpace(pname)); pname == "" {
			return
		}
		filesdata := params.filesdata
		if filesdata == nil {
			filesdata = snkio.NewMap[string, []File]()
			params.filesdata = filesdata
		}
		if fval, ok := filesdata.Load(pname); ok {
			if clear {
				fval = []File{}
				filesdata.Store(pname, fval)
			}
			if len(pfile) > 0 {
				for pf := range pfile {
					if fheader, fheaderok := pfile[pf].(*multipart.FileHeader); fheaderok {
						fval = append(fval, nextfile(fheader))
					}
				}
			}
			filesdata.Store(pname, fval)
		} else {
			if len(pfile) > 0 {
				var val = []File{}
				for pf := range pfile {
					if fheader, fheaderok := pfile[pf].(*multipart.FileHeader); fheaderok {
						val = append(val, nextfile(fheader))
					}
				}
				filesdata.Store(pname, val)
				params.filesdatacount++
			} else {
				filesdata.Store(pname, []File{})
				params.filesdatacount++
			}
		}
	}
}

// FileEmpty -> return true if no file parameters exist
func (params *parameters) FileEmpty(pname ...string) (empty bool) {
	if params != nil {
		empty = true
		filesdata := params.filesdata
		if filesdata == nil {
			return
		}
		pnl := len(pname)
		filesdata.Range(func(key string, value []File) bool {
			if pnl == 0 {
				empty = false
				return false
			}
			if pnl == 1 {
				if pname[0] != "" && strings.EqualFold(key, pname[0]) {
					for _, v := range value {
						if empty = v.Size() == 0; empty {
							continue
						}
						return false
					}
					return false
				}
				return empty
			}
			return false
		})
	}
	return empty
}

// FileExist -> check if file parameter exist
// pname : name
func (params *parameters) FileExist(pname string) bool {
	if params != nil {
		if pname = strings.ToUpper(strings.TrimSpace(pname)); pname == "" {
			return false
		}
		filesdata := params.filesdata
		if filesdata == nil {
			return false
		}
		_, ok := filesdata.Load(pname)
		return ok
	}
	return false
}

// RemoveFile  -> remove file parameter and return any slice of File value
func (params *parameters) RemoveFile(pname string) (value []File) {
	if pname = strings.ToUpper(strings.TrimSpace(pname)); pname == "" {
		return
	}
	filesdata := params.filesdata
	if filesdata == nil {
		return
	}
	if stdval, ok := filesdata.LoadAndDelete(pname); ok {
		params.filesdatacount--
		value = stdval
	}
	return
}

// Get - return a specific parameter values
func (params *parameters) Get(pname string, index ...int) []string {
	if params != nil {
		if pname = strings.ToUpper(strings.TrimSpace(pname)); pname != "" {
			if standard := params.standard; standard != nil {
				if stdval, ok := standard.Load(pname); ok {
					if stdl := len(stdval); stdl > 0 {
						if il := len(index); il > 0 {
							idx := []int{}
							for idn := range index {
								if id := index[idn]; id >= 0 && id < stdl {
									idx = append(idx, id)
								}
							}
							if len(idx) > 0 {
								stdvls := make([]string, len(idx))
								for in := range idx {
									stdvls[in] = stdval[idx[in]]
								}
								return stdvls
							}
						} else {
							return stdval
						}
					}
				}
			}
		}
	}
	return emptyParmVal
}

// String return parameter as string concatenated with sep
func (params *parameters) String(pname string, sep string, index ...int) (s string) {
	if params != nil {
		if pval := params.Get(pname, index...); len(pval) > 0 {
			return strings.Join(pval, sep)
		}
		if pval := params.GetFile(pname, index...); len(pval) > 0 {
			var rnrtos = func(br *bufio.Reader) (bs string, err error) {
				rns := make([]rune, 1024)
				rnsi := 0
				if br != nil {
					for {
						rn, size, rnerr := br.ReadRune()
						if size > 0 {
							rns[rnsi] = rn
							rnsi++
							if rnsi == len(rns) {
								bs += string(rns[:rnsi])
								rnsi = 0
							}
						}
						if rnerr != nil {
							if rnerr != io.EOF {
								err = rnerr
							}
							break
						}
					}
				}
				if rnsi > 0 {
					bs += string(rns[:rnsi])
					rnsi = 0
				}
				return
			}
			var bfr *bufio.Reader = nil
			for rn := range pval {
				if r := pval[rn]; r != nil {
					if bfr == nil {
						bfr = bufio.NewReader(r.Reader())
					} else {
						bfr.Reset(r.Reader())
					}
					if bfr != nil {
						if bs, bserr := rnrtos(bfr); bserr == nil {
							s += bs
						} else {
							break
						}
					}
				}
				if rn < len(pval)-1 {
					s += sep
				}
			}
		}
	}
	return
}

// GetFile return file paramater - array of file
func (params *parameters) GetFile(pname string, index ...int) []File {
	if params != nil {
		if pname = strings.ToUpper(strings.TrimSpace(pname)); pname != "" {
			filesdata := params.filesdata
			if filesdata != nil {
				if flsvv, ok := filesdata.Load(pname); ok {
					if flsl := len(flsvv); flsl > 0 {
						if il := len(index); il > 0 {
							idx := []int{}
							for _, id := range index {
								if id >= 0 && id < flsl {
									idx = append(idx, id)
								}
							}
							if len(idx) > 0 {
								flsvls := make([]File, len(idx))
								for in, id := range idx {
									flsvls[in] = flsvv[id]
								}
								return flsvls
							}
						} else {
							return flsvv
						}
					}
				}
			}
		}
	}
	return emptyParamFile
}

// Clear all standard parameters
func (params *parameters) Clear() {
	if params == nil {
		return
	}
	if standard, urlkeys := params.standard, params.urlkeys; standard != nil {
		params.standard = nil
		params.urlkeys = nil
		params.standardcount = 0
		standard.Range(func(key string, value []string) bool {
			if urlkeys != nil {
				urlkeys.LoadAndDelete(key)
			}
			_, delok := standard.LoadAndDelete(key)
			return !delok
		})
	}
}

// Clear all file parameters
func (params *parameters) ClearFiles() {
	if params == nil {
		return
	}
	if filesdata := params.filesdata; filesdata != nil {
		params.filesdata = nil
		params.filesdatacount = 0
		var delfls []File
		filesdata.Range(func(key string, value []File) bool {
			_, delok := filesdata.LoadAndDelete(key)
			delfls = append(delfls, value...)
			return !delok
		})
		for _, delf := range delfls {
			delf.Close()
		}
	}
}

// ClearAll function that can be called to assist in cleaning up instance of Parameter container
func (params *parameters) ClearAll() {
	if params == nil {
		return
	}
	params.Clear()
	params.ClearFiles()
}

// NewParameters return new instance of Paramaters container
func NewParameters() *parameters {
	return &parameters{}
}

// LoadParametersFromRawURL - populate paramaters just from raw url
func LoadParametersFromRawURL(params Parameters, rawURL string) (err error) {
	if params != nil && rawURL != "" {
		if rawURL != "" {
			rawURL = strings.Replace(rawURL, ";", "%3b", -1)
			var rawUrls = strings.Split(rawURL, "&")
			rawURL = ""
			for _, rwurl := range rawUrls {
				if rwurl != "" {
					if strings.Contains(rwurl, "=") {
						rawURL += rwurl + "&"
						continue
					}
					continue
				}
			}
			if len(rawURL) > 1 && strings.HasSuffix(rawURL, "&") {
				rawURL = rawURL[:len(rawURL)-1]
			}
			var urlvals url.Values
			if urlvals, err = url.ParseQuery(rawURL); err == nil {
				if len(urlvals) > 0 {
					for pname, pvalue := range urlvals {
						storeParameter(params.(*parameters), true, pname, false, pvalue...)
					}
				}
			}
		}
	}
	return
}

// LoadParametersFromUrlValues - Load Parameters from url.Values
func LoadParametersFromUrlValues(params Parameters, urlvalues url.Values) (err error) {
	if params != nil && urlvalues != nil {
		for pname, pvalue := range urlvalues {
			params.Set(pname, false, pvalue...)
		}
	}
	return
}

// LoadParametersFromMultipartForm - Load Parameters from *multipart.Form
func LoadParametersFromMultipartForm(params Parameters, mpartform *multipart.Form) (err error) {
	if params != nil && mpartform != nil {
		for pname, pvalue := range mpartform.Value {
			params.Set(pname, false, pvalue...)
		}
		for pname, pfile := range mpartform.File {
			if len(pfile) > 0 {
				pfilei := []interface{}{}
				for _, pf := range pfile {
					pfilei = append(pfilei, pf)
				}
				params.SetFile(pname, false, pfilei...)
				pfilei = nil
			}
		}
	}
	return
}

func LoadParametersFromReader(params Parameters, rdr io.Reader) (err error) {
	var bts []byte
	var btsl int
	var rb = make([]byte, 1)
	rn := 0
	for {
		if rn, err = rdr.Read(rb); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		if rn > 0 {
			bts = append(bts, rb...)
			if btsl++; btsl == 2 && string(bts) == "--" {
				for {
					if rn, err = rdr.Read(rb); err != nil {
						if err == io.EOF {
							err = nil
						}
						return
					}
					if rn > 0 {
						bts = append(bts, rb...)
						if btsl++; btsl > 1 {
							if string(bts[btsl-2:btsl]) == "\r\n" {
								var mltr *multipart.Reader
								var rawrdr = snkio.ReadFunc(func(p []byte) (n int, err error) {
									pl := len(p)
									for n < pl {
										if btsl > 0 {
											if btsl <= (pl - n) {
												c := copy(p[n:], bts)
												n += c
												bts = nil
												btsl = 0
												return
											}
											c := copy(p[n:], bts[:(pl-n)])
											n += c
											if btsl -= c; btsl > 0 {
												continue
											}
											return
										}
										rn, rerr := rdr.Read(p[n:])
										n += rn
										err = rerr
										return
									}
									if n == 0 {
										err = io.EOF
									}
									return
								})
								if string(bts[btsl-4:btsl-2]) == "--" {
									mltr = multipart.NewReader(rawrdr, string(bts[2:btsl-4]))
								}
								if mltr == nil {
									mltr = multipart.NewReader(rawrdr, string(bts[2:btsl-2]))
								}
								frm, frmerr := mltr.ReadForm(0)
								if frmerr != nil {
									if frmerr == io.EOF {
										return nil
									}
									return frmerr
								}
								return LoadParametersFromMultipartForm(params, frm)
							}
						}
						continue
					}
					break
				}
			}
			continue
		}
		break
	}
	return
}

func HTTPRequestParameters(r *http.Request) (params Parameters, err error) {
	params = NewParameters()
	if r.URL != nil {
		if err = LoadParametersFromRawURL(params, r.URL.RawQuery); err != nil {
			params.ClearAll()
			params = EmptyParameters
			return
		}
		r.URL.RawQuery = ""
	}
	if err = r.ParseMultipartForm(0); err == nil {
		if r.MultipartForm != nil {
			if err = LoadParametersFromMultipartForm(params, r.MultipartForm); err != nil {
				params.ClearAll()
				params = EmptyParameters
				return
			}
			return
		}
		if r.Form != nil {
			if err = LoadParametersFromUrlValues(params, r.Form); err != nil {
				params.ClearAll()
				params = EmptyParameters
				return
			}
			return
		}
		return
	}
	if err = r.ParseForm(); err == nil {
		if err = LoadParametersFromUrlValues(params, r.Form); err == nil {
			return
		}
	}
	params.ClearAll()
	params = EmptyParameters
	return
}

// LoadParametersFromHTTPRequest - Load Parameters from http.Request
func LoadParametersFromHTTPRequest(params Parameters, r *http.Request) (err error) {
	if params != nil {
		if r.URL != nil {
			if err = LoadParametersFromRawURL(params, r.URL.RawQuery); err != nil {
				return
			}
			r.URL.RawQuery = ""
		}
		if err = r.ParseMultipartForm(0); err == nil {
			if r.MultipartForm != nil {
				LoadParametersFromMultipartForm(params, r.MultipartForm)
			} else if r.Form != nil {
				LoadParametersFromUrlValues(params, r.Form)
			}
		} else if err = r.ParseForm(); err == nil {
			LoadParametersFromUrlValues(params, r.Form)
		}
	}
	return
}
