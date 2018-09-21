package hasher

import "testing"

func TestHash(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"angryMonkey", "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
	}
	for _, c := range cases {
		got := EncodeSha512Base64(c.in)
		if got != c.want {
			t.Errorf("Hashed %q wrong. Expect %q Got %q", c.in, c.want, got)
		}
	}

}
