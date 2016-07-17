package main

import (
	"fmt"
	"log"
  "bufio"
  "os"
  "bytes"
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

func main() {
  var buffer bytes.Buffer

  scanner := bufio.NewScanner(os.Stdin)
  scanner.Split(bufio.ScanBytes)

  for scanner.Scan() {
    buffer.WriteString(scanner.Text())
  }

  source := buffer.String()

  yaspPeg := &YaspPEG{Buffer: source}
  yaspPeg.Init()
  yaspPeg.Parsing.Init()
  if err := yaspPeg.Parse(); err != nil {
    log.Fatal(err)
  }
  yaspPeg.Execute()

  context := EmptyEvaluationContext()

  yaspPeg.Evaluate(context)

  fmt.Printf("%v\n", context.vars["main"].EvaluateFunction(context, []*Value{&Value{T: TypeString, V: "  (+ 1 2)"}}))

  fmt.Print("done\n")
}
