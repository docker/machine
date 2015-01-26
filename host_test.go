package main

import (
	"testing"
)

func TestValidateHostnameValid(t *testing.T) {
	hosts := []string{
		"zomg",
		"test-ing",
		"some.h0st",
	}

	for _, v := range hosts {
		h, err := ValidateHostName(v)
		if err != nil {
			t.Fatal("Invalid hostname")
		}

		if h != v {
			t.Fatal("Hostname doesn't match")
		}
	}
}

func TestValidateHostnameInvalid(t *testing.T) {
	hosts := []string{
		"zom_g",
		"test$ing",
		"some😄host",
	}

	for _, v := range hosts {
		_, err := ValidateHostName(v)
		if err == nil {
			t.Fatal("No error returned")
		}
	}
}
