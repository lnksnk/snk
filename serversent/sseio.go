package serversent

import (
	"fmt"
	"io"

	snkio "github.com/lnksnk/snk/io"
)

func SseFPrint(ssew io.Writer, id, event string, a ...any) (err error) {
	var hasid, hasevt, hasdata, didprnt = id != "", event != "", len(a) > 0, false
	if hasid {
		if err = writesseid(ssew, []byte(id)...); err != nil {
			return
		}
		didprnt = true
	}
	if err == nil && hasevt {
		if err = writesseevent(ssew, []byte(event)...); err != nil {
			return
		}
		didprnt = true
	}
	if hasdata {
		var sb []byte
		var si = 0
		var ssewtr = snkio.WriteFunc(func(p []byte) (n int, swerr error) {
			for pi := range len(p) {
				if p[pi] == '\n' {
					if swerr = writessedata(ssew, sb...); swerr == nil {
						sb = nil
						si = 0
						continue
					}
					return
				}
				n++
				sb = append(sb, p[pi])
				if si++; si == maxdatalen {
					if swerr = writessedata(ssew, sb...); swerr == nil {
						sb = nil
						si = 0
						continue
					}
					return
				}
			}
			return
		})
		if _, err = snkio.Fprint(ssewtr, a...); err != nil {
			return
		}
		if si > 0 {
			if err = writessedata(ssewtr, sb...); err != nil {
				return
			}
			sb = nil
			si = 0
		}
		didprnt = true
	}
	if err == nil {
		if didprnt {
			err = writessenl(ssew)
			return
		}
		return fmt.Errorf("No id, event or data")
	}
	return
}

var sseid = []byte("id: ")
var sseevent = []byte("event: ")
var ssedata = []byte("data: ")
var ssenl = []byte{'\n'}

func writessenl(sscn io.Writer) (dtaerr error) {
	_, dtaerr = sscn.Write(ssenl)
	return
}

func writesseid(sscn io.Writer, b ...byte) (dtaerr error) {
	if _, dtaerr = sscn.Write(sseid); dtaerr == nil {
		if _, dtaerr = sscn.Write(b); dtaerr == nil {
			return writessenl(sscn)
		}
		return
	}
	return
}

func writesseevent(sscn io.Writer, b ...byte) (dtaerr error) {
	if _, dtaerr = sscn.Write(sseevent); dtaerr == nil {
		if _, dtaerr = sscn.Write(b); dtaerr == nil {
			return writessenl(sscn)
		}
		return
	}
	return
}

func writessedata(sscn io.Writer, b ...byte) (dtaerr error) {
	if _, dtaerr = sscn.Write(ssedata); dtaerr == nil {
		if _, dtaerr = sscn.Write(b); dtaerr == nil {
			return writessenl(sscn)
		}
		return
	}
	return
}
