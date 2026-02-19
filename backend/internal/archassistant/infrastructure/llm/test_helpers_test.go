package llm_test

import "io"

func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
