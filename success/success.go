package success

import (
	`testing`
)

type Success struct {
	T *testing.T
	E error
	A any
	M string
}

func (s *Success) SetT(t *testing.T) *Success {
	s.T = t
	return s
}

func (s *Success) SetE(e error) *Success {
	s.E = e
	return s
}

func (s *Success) SetA(a any) *Success {
	s.A = a
	return s
}

func (s *Success) SetM(m string) *Success {
	s.M = m
	return s
}

func (s *Success) ExpectError() bool {
	return s.E != nil
}

func (s *Success) ExpectNoError() bool {
	return s.E == nil
}

func (s *Success) Result() bool {
	if s.E != nil {
		if s.T != nil {
			s.T.Errorf("%s ; received error %v", s.M, s.E)
		}
		return false
	}
	return true
}

func (s *Success) AnyError() (any, error) {
	return s.A, s.E
}

func (s *Success) Evaluate() (any, bool) {
	return s.A, s.Result()
}

func Succeeded(e error) *Success {
	return &Success{E: e}
}

func AnySucceeded(a any, e error) *Success {
	return &Success{
		T: nil,
		E: e,
		A: a,
		M: "",
	}
}
