package io

type WriteRunes interface {
	WriteRunes(...rune) error
}

type WriteRunesFunc func(...rune) error

func (wtrrnsfnc WriteRunesFunc) WriteRunes(r ...rune) error {
	return wtrrnsfnc(r...)
}
