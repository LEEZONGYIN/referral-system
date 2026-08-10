package app

import "testing"

func TestNewHTTPHandler(t *testing.T) {
	_ = NewHTTPHandler(nil)
}
