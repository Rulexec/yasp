package yaspUtil

import (
  "log"

  . "../src"
)

func ParseAndEvaluate(expr string) *Value {
  yaspPeg := &YaspPEG{Buffer: expr}
  yaspPeg.Init()
  yaspPeg.Parsing.Init()
  if err := yaspPeg.Parse(); err != nil {
    log.Fatal(err)
  }
  yaspPeg.Execute()
	return yaspPeg.Evaluate(EmptyEvaluationContext())
}
