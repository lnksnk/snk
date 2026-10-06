package io

type WriteFunc func([]byte) (n int, err error)

func (wfnc WriteFunc) Write(p []byte) (n int, err error) {
	return wfnc(p)
}
