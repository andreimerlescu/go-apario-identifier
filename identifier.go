package go_apario_identifier

import (
	`fmt`
)

type Identifier struct {
	Year int16  `json:"y"`
	Code string `json:"c"`
}

func (i *Identifier) Identifier() *Identifier {
	return i
}

func (i *Identifier) ID() string {
	return i.String()
}

func (i *Identifier) Path() string {
	return IdentifierPath(i.String())
}

func (i *Identifier) String() string {
	return fmt.Sprintf("%04d%s", i.Year, i.Code)
}
