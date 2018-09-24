package hasher

import "testing"

func TestHashSuccess(t *testing.T) {
	cases := []struct {
		passwsd, hash string
	}{
		{"angryMonkey", "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
	}
	for _, c := range cases {
		result := EncodeSha512Base64(c.passwsd)
		if result != c.hash {
			t.Errorf("Hashed %q wrong. Expect %q Got %q", c.passwsd, c.hash, result)
		}
	}

}

func TestHashFailure(t *testing.T) {
	cases := []struct {
		passwsd, hash string
	}{
		{"AngryMonkey", "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
	}
	for _, c := range cases {
		result := EncodeSha512Base64(c.passwsd)
		if result == c.hash {
			t.Errorf("Hashed %q wrong. Should not be %q", c.passwsd, c.hash)
		}
	}

}
