package yasp

import (
  "testing"

  //. "../src";
  . "../util";
)

func TestStruct(t *testing.T) {
  var expected uint64 = 3
  actual := ParseAndEvaluate("(defstruct Some a b)\n((Some 2 3) b)")

  AssertNumber(t, expected, actual)
}
