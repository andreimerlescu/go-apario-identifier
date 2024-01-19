package go_apario_identifier

import (
	`context`
	`errors`
	`log`
	`os`
	`path/filepath`
	`strconv`
	`strings`
	`time`
)

func NewIdentifier(databasePrefixPath string, identifierLength int, attemptsCounter int, timeoutSeconds int) (*Identifier, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
	defer cancel()

	ticker := time.NewTicker(33 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			identifier, identifierErr := generateIdentifier(databasePrefixPath, identifierLength, attemptsCounter)
			if identifierErr != nil {
				log.Printf("failed to acquire new identifier with err: %v", identifierErr)
			} else {
				if len(identifier.String()) > 6 {
					return identifier, nil
				}
			}

		}
	}
}

func ParseIdentifier(identifier string) (*Identifier, error) {
	identifier = strings.ToLower(identifier)
	tsIdentifier := &Identifier{
		Year: 0,
		Code: "",
	}
	var yearString string = identifier[0:4]
	var codeString string = identifier[4:]
	year, intErr := strconv.Atoi(yearString)
	if intErr != nil {
		return nil, intErr
	}

	tsIdentifier.Code = codeString
	tsIdentifier.Year = int16(year)

	// in order for the directory to be considered a year, it must be +/- 17 years from the current date. this value should be configurable
	years := time.Duration(17)
	if !(year > time.Now().UTC().AddDate(int(-1*years), 0, 0).Year() && year < time.Now().UTC().Add(years*13*28*24*time.Hour).Year()) {
		return tsIdentifier, errors.New("identifier year octet out of range of +/- 17 years from time.Now()")
	}

	return tsIdentifier, nil
}

func IdentifierPath(identifier string) string {
	identifier = strings.ToLower(identifier)
	debug := len(os.Getenv("DEBUG")) > 0
	var paths []string
	var depth, prev, remaining int = 1, 0, 0
	for {
		if prev == 0 {
			prev = 4
			if prev > len(identifier) {
				paths = append(paths, identifier)
				break
			}
			paths = append(paths, identifier[0:prev])
			remaining = len(identifier) - prev
			continue
		}

		r := fibonacci(depth)
		if r > remaining {
			r = remaining
		}
		left := prev
		right := r + prev
		if right >= len(identifier) {
			right = len(identifier)
		}
		if debug {
			log.Printf("[DEBUG] r = %d for depth = %d ; left = %d ; right = %d ; len = %d", r, depth, left, right, len(identifier))
		}
		paths = append(paths, identifier[left:right])
		if right >= len(identifier) {
			break
		}
		prev = right
		depth++
	}
	return filepath.Join(paths...)
}
