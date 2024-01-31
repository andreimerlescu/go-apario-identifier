package go_apario_identifier

import (
	`crypto/rand`
	`errors`
	`fmt`
	`log`
	`math/big`
	`os`
	`path/filepath`
	`time`
)

// newToken this is attempts squared with a length of the token
func newToken(length int, attempts int) (*Identifier, error) {
	if length < 0 {
		return nil, errors.New("token length must be > 0")
	}
	if attempts < 1 {
		return nil, errors.New("no remaining attempts left")
	}
	if length > 30 {
		return nil, errors.New("maximum token length is 29 chars")
	}

	for {
		token := make([]byte, length)
		for i := range token {
			m := big.NewInt(int64(len(IdentifierCharset)))
			randIndex, err := rand.Int(rand.Reader, m)
			if err != nil {
				log.Printf("failed to generate random number: %v", err)
				continue
			}
			token[i] = IdentifierCharset[randIndex.Int64()]
		}

		id := fmt.Sprintf("%4d%v", time.Now().UTC().Year(), string(token))

		identifier, identifierErr := ParseIdentifier(id)
		if identifierErr != nil {
			attempts += 1
			if attempts <= 17 {
				return newToken(length, attempts)
			}
			return nil, errors.New("failed to generate acceptable token after 17 attempts")
		}
		return identifier, nil
	}
}

// generateIdentifier takes the databasePrefixPath and a newToken(length, attempts) and attempts an os.MkdirAll on the
// filesystem to verify whether or not the identifier currently exists. The newToken(length, attempts) result is
// converted to a filepath and then verified using an os.Stat. If the identifier/path exists, then this func is
// recursively called up to the remaining attempts > 0. This func uses a semaphore that limits the concurrent runtime
// of the func to 1. This ensures that there are not more than 1 attempt being made to write to the database prefix.
func generateIdentifier(databasePrefixPath string, length int, attempts int) (*Identifier, error) {
	valet := NewValet(databasePrefixPath)
	valet.SafetyCheck()
	var identifier string
	cache, cacheErr := valet.GetCache(databasePrefixPath)
	if cacheErr != nil {
		return nil, cacheErr
	}
	attemptedIdentifier, attemptErr := newToken(length, attempts)
	if attemptErr != nil {
		return nil, attemptErr
	}
	identifierPath := IdentifierPath(attemptedIdentifier.String())
	path := filepath.Join(databasePrefixPath, identifierPath)
	_, infoErr := os.Stat(path)
	if infoErr != nil {
		// error with path, lets create it
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return nil, err
		}
		err = cache.Write(attemptedIdentifier.String(), 1)
		if err != nil {
			log.Printf("failed to cache.Write(%v, 1) due to err %v", identifier, err)
			return nil, err
		}
		return attemptedIdentifier, nil
	} else {
		log.Printf("[retrying] identifier exists at path: %v", path)
		attempts += 1
		if attempts <= 17 {
			return generateIdentifier(databasePrefixPath, length, attempts)
		}
		return nil, errors.New("failed to acquire new unique identifier within allotted attempt window of opportunity")
	}
}
