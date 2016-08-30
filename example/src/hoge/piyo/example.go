package piyo

import (
	"github.com/kai-zoa/geeyoko"
)

type Geeyoko int

func (Geeyoko) PostalCode() string {
	return geeyoko.PostalCode()
}

func (Geeyoko) StationAddress() string {
	return geeyoko.StationAddress()
}
