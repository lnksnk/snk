package serve

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func RequestRangeInfo(hddr http.Header) (rangeOffset int64, rangeType string) {
	rangeType = ""
	rangeOffset = int64(-1)
	if prtclrange := hddr.Get("Range"); prtclrange != "" && strings.Index(prtclrange, "=") > 0 {
		if rangeType = prtclrange[:strings.Index(prtclrange, "=")]; prtclrange != "" {
			if prtclrange = prtclrange[strings.Index(prtclrange, "=")+1:]; strings.Index(prtclrange, "-") > 0 {
				rangeOffset, _ = strconv.ParseInt(prtclrange[:strings.Index(prtclrange, "-")], 10, 64)
			}
		}
	}
	return
}

func ServeHttpRange(w http.ResponseWriter, r *http.Request, source io.ReadSeeker, srcsize int64) (err error) {
	var rangeOffset, rangeType = RequestRangeInfo(r.Header)
	if rangeType == "" && rangeOffset == -1 && source != nil {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", srcsize))
		w.WriteHeader(http.StatusOK)
		_, err = io.CopyN(w, source, srcsize)
		return
	}
	if rangeType != "" && rangeOffset > -1 && source != nil {
		if srcsize > 0 {
			source.Seek(rangeOffset, 0)
			maxoffset := int64(0)
			maxlen := int64(0)
			if maxoffset = rangeOffset + (srcsize - rangeOffset); maxoffset > 0 {
				maxlen = maxoffset - rangeOffset
				maxoffset--
			}

			if maxoffset < rangeOffset {
				maxoffset = rangeOffset
				maxlen = 0
			}
			contentrange := fmt.Sprintf("%s %d-%d/%d", rangeType, rangeOffset, maxoffset, srcsize)
			w.Header().Set("Content-Range", contentrange)
			w.Header().Set("Content-Length", fmt.Sprintf("%d", maxlen))
			w.WriteHeader(http.StatusPartialContent)
			if _, err = source.Seek(rangeOffset, io.SeekStart); err == nil {
				_, err = io.CopyN(w, source, maxlen)
			}
			return
		}
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", 0))
	return
}
