package main

import ( "fmt" )

type Type uint8
const (
  TypeID Type = iota
  TypeNumber
  TypeExpression
  TypeFunction
  TypeNil
)

type Value struct {
  T Type
  V interface {}
}

type ValueFunction struct {
  boundContext *EvaluationContext
  argsNames []string
  body *Value
}

func (v *Value) String() string {
  switch v.T {
  case TypeID: {
    s, _ := v.V.(string)
    return s
  }
  case TypeNumber: {
    n, _ := v.V.(uint64)
    return fmt.Sprint(n)
  }
  case TypeExpression: {
    s, _ := v.V.(*Stack)
    return s.String()
  }
  default: fmt.Printf(" %v ", v.T); panic("unknown type")
  }
}
