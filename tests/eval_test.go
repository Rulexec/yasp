package yasp

import (
  "testing"

  //. "../src";
  . "../util";
)

func TestEvalSum(t *testing.T) {
  var expected uint64 = 2 + 3
  actual := ParseAndEvaluate("(eval (YaspNode LIST (list (YaspNode ID '+') (YaspNode NUMBER 2) (YaspNode NUMBER 3))))")

  AssertNumber(t, expected, actual)
}
