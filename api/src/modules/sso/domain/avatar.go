package domain

import (
	"io"
)

type AvatarFile struct {
	Filename  string
	Extension string

	Data io.Reader
}
