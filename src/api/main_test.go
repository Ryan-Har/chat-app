package main

import (
	"testing"
)

func TestVerifyTimeFormat(t *testing.T) {
	t.Run("yyyy-MM-dd HH:mm:ss.SSSSSS format", func(t *testing.T) {
		if _, err := verifyTimeFormat("2024-02-27 15:35:20.311231"); err != nil {
			t.Error("error validating with this format")
		}
	})

	t.Run("yyyy-MM-dd HH:mm:ss.SSS format", func(t *testing.T) {
		if _, err := verifyTimeFormat("2024-02-27 15:35:20.311"); err != nil {
			t.Error("error validating with this format")
		}
	})
}
