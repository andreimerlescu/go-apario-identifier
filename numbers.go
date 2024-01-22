package go_apario_identifier

import (
	`fmt`
	`strings`
	`time`
)

type Fragment []rune

func CodeFragment(identifier string) Fragment {
	identifier = strings.ToUpper(identifier)
	return Fragment(identifier) // return the fragment
}

func IdentifierFragment(identifier string) Fragment {
	if len(identifier) > 4 {
		identifier = identifier[4:] // strip out the year
	}
	return CodeFragment(identifier) // return the fragment
}

func IntegerFragment(num int) Fragment {
	return Fragment(EncodeBase36(num))
}

func (f Fragment) String() string {
	return string(f)
}

func (f Fragment) ToIdentifier() (*Identifier, error) {
	return f.ToYearIdentifier(time.Now().UTC().Year())
}

func (f Fragment) ToYearIdentifier(year int) (*Identifier, error) {
	return ParseIdentifier(fmt.Sprintf("%4d%s", year, f.String()))
}
