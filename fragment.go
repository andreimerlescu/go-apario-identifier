package go_apario_identifier

import (
	`fmt`
	`path/filepath`
	`strings`
	`sync/atomic`
	`time`
)

type Fragment []rune

func CodeFragment(code string) Fragment {
	code = strings.ToUpper(code)
	return Fragment(code) // return the fragment
}

func IntegerFragment(num int) Fragment {
	return Fragment(EncodeBase36(num))
}

func (f Fragment) String() string {
	return string(f)
}

func (f Fragment) Path() string {
	identifier := strings.ToUpper(string(f))
	var paths []string
	depth := atomic.Int32{}
	prev := atomic.Int32{}
	remaining := atomic.Int32{}
	depth.Store(0)
	prev.Store(0)
	remaining.Store(int32(len(identifier)))
	for {
		// do the task
		r := rFibonacci(int(depth.Add(1)))
		if int32(r) > remaining.Load() {
			r = int(remaining.Load())
		}
		left := int(prev.Load())
		right := r + int(prev.Load())
		if right >= len(identifier) {
			right = len(identifier)
		}
		segment := identifier[left:right]

		// save the result
		if len(segment) == 0 {
			break
		}
		paths = append(paths, segment)

		// prepare the next loop
		prev.Store(int32(right))
		remaining.Store(int32(len(identifier[right:])))

		// when to break
		if remaining.Load() < 0 {
			break
		}
	}
	return filepath.Join(paths...)
}

func (f Fragment) ToIdentifier() (*Identifier, error) {
	return f.ToYearIdentifier(time.Now().UTC().Year())
}

func (f Fragment) ToYearIdentifier(year int) (*Identifier, error) {
	return ParseIdentifier(fmt.Sprintf("%4d%s", year, f.String()))
}
