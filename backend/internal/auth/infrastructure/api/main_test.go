package api

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_ = os.Unsetenv("AUTH_MODE")
	os.Exit(m.Run())
}
