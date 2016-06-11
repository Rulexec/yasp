package main

import (
	"fmt"
	"log"
)

func parseAndEvaluate(expr string) *Value {
  yaspPeg := &YaspPEG{Buffer: expr}
  yaspPeg.Init()
  yaspPeg.Parsing.Init()
  if err := yaspPeg.Parse(); err != nil {
    log.Fatal(err)
  }
  yaspPeg.Execute()
	return yaspPeg.Evaluate()
}

func main() {
  //yaspPeg := &YaspPEG{Buffer: "(+ (* 2 2) 2)"}
  //expr := "(if 42 (print 13) (print 79))"
  expr := "(print 42) (print 78)"

	fmt.Printf("%v = %v\n", expr, parseAndEvaluate(expr))
}
