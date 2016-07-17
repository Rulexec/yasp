package main

import (
  . "..";
  "testing"
)

func TestAddition(t *testing.T) {
  var expected uint64 = 1 + 2 + 3
  actual := ParseAndEvaluate("(+ 1 (+ 2 3))")

  AssertNumber(t, expected, actual)
}

func TestFn(t *testing.T) {
  var expected uint64 = 6
  actual := ParseAndEvaluate("((fn (a b) (+ a b 1)) 2 3)")

  AssertNumber(t, expected, actual)
}

func TestDefn(t *testing.T) {
  var expected uint64 = 6
  actual := ParseAndEvaluate("(defn my-sum (a b) (+ a b 1))\n(my-sum 2 3)")

  AssertNumber(t, expected, actual)
}

func TestRecursion(t *testing.T) {
  var expected uint64 = 10
  actual := ParseAndEvaluate("((defn f (a b) (if (< a 5) (f (+ a 1) (+ b 2)) b)) 0 0)")

  AssertNumber(t, expected, actual)
}

func TestString(t *testing.T) {
  var expected string = "Hello, World!"
  actual := ParseAndEvaluate("(concat 'Hello, ' 'World!')")

  AssertString(t, expected, actual)
}

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
