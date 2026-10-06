package io

type Printer interface {
	Print(...any) error
	Println(...any) error
}

type PrintFunc func(...any) error

func (prntfnc PrintFunc) Print(a ...any) error {
	return prntfnc(a...)
}

type PrintlnFunc func(...any) error

func (prntlnfnc PrintFunc) Println(a ...any) error {
	return prntlnfnc(a...)
}
