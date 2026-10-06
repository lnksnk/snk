package conf

import "io"

type Config interface {
	Load(io.Reader) error
	LoadSection(string, io.Reader)
}

type ConfigSection interface {
}

func LoadConfig(io.Reader)
