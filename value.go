package main

import (
  "fmt"
  "bytes"
)

type Type uint8
const (
  TypeID Type = iota
  TypeNumber
  TypeString
  TypeExpression
  TypeFunction
  TypeNil
  TypeList
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
  case TypeString: {
    s, _ := v.V.(string)
    return "'" + s + "'"
  }
  case TypeNumber: {
    n, _ := v.V.(uint64)
    return fmt.Sprint(n)
  }
  case TypeExpression: {
    s, _ := v.V.(*Stack)
    return s.String()
  }
  case TypeFunction: {
    return "<function>"
  }
  case TypeList: {
    var buffer bytes.Buffer

    buffer.WriteString("[")

    lst, _ := v.V.([]*Value)
    size := len(lst)

    if size > 0 {
      buffer.WriteString(lst[0].String())

      for _, x := range lst[1:] {
        buffer.WriteString(", ")
        buffer.WriteString(x.String())
      }
    }

    buffer.WriteString("]")

    return buffer.String()
  }
  case TypeNil: {
    return "()"
  }
  default: panic(fmt.Sprintf("unknown type: %v", v.T))
  }
}

func (a *Value) Equals(b *Value) bool {
  if a.T != b.T { return false }

  switch a.T {
  case TypeString: {
    as, _ := a.V.(string)
    bs, _ := b.V.(string)

    return as == bs
  }
  default: panic(fmt.Sprintf("Unsupported type for comparison: %v", a.T))
  }
}

func (v *Value) Bool() bool {
  switch v.T {
  case TypeNumber: {
    n, _ := v.V.(uint64)
    return n != 0
  }
  default: panic(fmt.Sprintf("Unsupported type for boolean coersion: %v", v.T))
  }
}
