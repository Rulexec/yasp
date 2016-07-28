package main

import (
	"fmt"
	"log"
  "bufio"
  "os"
  "bytes"

  . "../.."
  //. "./util"
)

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

  fmt.Printf("%v\n", context.Vars["main"].EvaluateFunction(context, []*Value{&Value{T: TypeString, V: "  (+ 1 2)"}}))

  fmt.Print("done\n")
}
