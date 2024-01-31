package go_apario_identifier

import (
	`context`
	`log`
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
				return nil, identifierErr
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
		Year:     0,
		Fragment: Fragment{},
	}
	var yearString string = identifier[0:4]
	var codeString string = identifier[4:]
	year, intErr := strconv.Atoi(yearString)
	if intErr != nil {
		return nil, intErr
	}

	tsIdentifier.Fragment = CodeFragment(codeString)
	tsIdentifier.Year = int16(year)

	return tsIdentifier, nil
}

func IdentifierPath(identifier string) string {
	identifier = strings.ToUpper(identifier)
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

		r := rFibonacci(depth)
		if r > remaining {
			r = remaining
		}
		left := prev
		right := r + prev
		if right >= len(identifier) {
			right = len(identifier)
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
