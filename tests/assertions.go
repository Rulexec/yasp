package yasp

import (
  . "../src";
  "testing"
)

func AssertNumber(t *testing.T, expected uint64, actual *Value) {
  if actual.T != TypeNumber {
    t.Error("Expected TypeNumber, got: ", actual.T)
  }

  n, _ := actual.V.(uint64)

  if n != expected {
    t.Error("Expected ", expected, ", got: ", n)
  }
}

func AssertString(t *testing.T, expected string, actual *Value) {
  if actual.T != TypeString {
    t.Error("Expected TypeString, got: ", actual.T)
  }

  s, _ := actual.V.(string)

  if s != expected {
    t.Error("Expected ", expected, ", got: ", s)
  }
}
