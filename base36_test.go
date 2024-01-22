package go_apario_identifier

import (
	`testing`
)

func TestBase36(t *testing.T) {
	for i := 0; i < 369; i++ {
		encoded := EncodeBase36(i)
		decoded, err := DecodeBase36(encoded)
		if err != nil {
			t.Errorf("DecodeBase36(%d) received err %v", i, err)
			return
		}
		if decoded != i {
			t.Errorf("decoded %v != %v", decoded, i)
			return
		}
	}

}
