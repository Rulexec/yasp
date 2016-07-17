package yasp

import (
  "testing"

  //. "../src";
  . "../util";
)

func TestEnumCompare(t *testing.T) {
  var expected uint64 = 1
  actual := ParseAndEvaluate("(defenum Color RED GREEN BLUE)\n(= (Color RED) (Color RED))");

  AssertNumber(t, expected, actual)

  expected = 0
  actual = ParseAndEvaluate("(defenum Color RED GREEN BLUE)\n(= (Color RED) (Color GREEN))");

  AssertNumber(t, expected, actual)
}

func TestEnumSwitch(t *testing.T) {
  var expected uint64 = 2
  actual := ParseAndEvaluate("(defenum Color RED GREEN BLUE)\n(switch (Color GREEN) RED 1 GREEN 2)");

  AssertNumber(t, expected, actual)
}
