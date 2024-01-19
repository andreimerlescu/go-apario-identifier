package go_apario_identifier

type Identification interface {
	Identifier() *Identifier
	ID() string
	String() string
	Path() string
}

type ID Identification

func NewID(databasePath string, length int) (ID, error) {
	identifier, err := NewIdentifier(databasePath, length, 17, 30)
	if err != nil {
		return nil, err
	}
	return identifier, nil
}
