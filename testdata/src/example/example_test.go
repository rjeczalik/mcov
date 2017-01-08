package example_test

import "testing"

func TestExample(t *testing.T) {
	cases := []struct {
		fun  string
		arg  string
		want string
	}{{
		"echo",
		"hello world",
		"hello world",
	}, {
		"reverse",
		"hello world",
		"dlrow olleh",
	}}

	for _, cas := range cases {
		t.Run(cas.fun+" "+cas.arg, func(t *testing.T) {
			got, err := call(cas.fun, cas.arg)
			if err != nil {
				t.Fatalf("call()=%s", err)
			}

			if got != cas.want {
				t.Fatalf("got %q, want %q", got, cas.want)
			}
		})
	}
}
