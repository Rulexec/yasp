package main

import (
  "fmt"
)

func (p *Parsing) Evaluate() *Value {
  if (p.stack.IsEmpty()) {
    return nil
  }

  //return p.stack.Evaluate(EmptyEvaluationContext())
  expanded := p.stack.Expand()

  context := EmptyEvaluationContext()
  var last *Value

  for _, x := range expanded {
    last = x.Evaluate(context)
  }

  return last
}

func CreateFunction(context *EvaluationContext, argsStack *Stack, body *Value) *Value {
  argsExpanded := argsStack.Expand()

  argsNames := make([]string, argsStack.Size)

  for i, x := range argsExpanded {
    if x.T != TypeID { panic("list in second function argument must contain only ids") }

    str, _ := x.V.(string)
    argsNames[i] = str
  }

  return &Value{T: TypeFunction, V: ValueFunction{boundContext: context, argsNames: argsNames, body: body}}
}

func AssertNumberOfArguments(s *Stack, expected uint32, fnName string) {
  if s.Size - 1 != expected {
    panic(fmt.Sprintf("%v expected %v args, got %v", fnName, expected, s.Size - 1))
  }
}
func (v *Value) AssertNumberType() uint64 {
  if v.T != TypeNumber { panic(fmt.Sprintf("%v expected to be Number", v)) }

  n, _ := v.V.(uint64)

  return n
}
func (s *Stack) Evaluate(context *EvaluationContext) *Value {
  if s.Size == 0 { return &Value{T: TypeNil} }

  expanded := s.Expand()

  f := expanded[0].Evaluate(context)

  if (s.Prev == nil && s.Size == 1) { return f }

  switch f.T {
  case TypeNumber: fmt.Printf(" %v ", s); panic("Number is not a function")
  case TypeID: {
    fnName, _ := f.V.(string)
    switch fnName {
    case "+": {
     var result uint64 = 0

     for _, x := range expanded[1:] {
       xv := x.Evaluate(context)
       if (xv.T != TypeNumber) { panic("is not a number for +") }

       n, _ := xv.V.(uint64)
       result += n
     }

     return &Value{T: TypeNumber, V: result}
    }
    case "<": {
      AssertNumberOfArguments(s, 2, fnName)

      a, b := expanded[1].Evaluate(context).AssertNumberType(),
              expanded[2].Evaluate(context).AssertNumberType()

      if a < b {
        return &Value{T: TypeNumber, V: uint64(1)}
      } else {
        return &Value{T: TypeNumber, V: uint64(0)}
      }
    }
    case "fn": {
      if s.Size != 3 { fmt.Printf(" %v ", s); panic("function must have 2 args") }

      if expanded[1].T != TypeExpression { panic("second function arguments must be a list") }

      argsStack, _ := expanded[1].V.(*Stack)
      return CreateFunction(context, argsStack, expanded[2])
    }
    case "print": {
     for _, x := range expanded[1:] {
       xv := x.Evaluate(context)
       fmt.Println(xv.V)
     }

     return &Value{T: TypeNumber, V: uint64(1)}
    }
    case "if": {
      if s.Size != 4 { panic("if must have 3 args") }

      condition := expanded[1].Evaluate(context)
      switch condition.T {
      case TypeNumber: {
        n, _ := condition.V.(uint64)

        var branch *Value

        if n != 0 { branch = expanded[2]
        }    else { branch = expanded[3] }

        return branch.Evaluate(context)
      }
      default: panic(fmt.Sprintf("Unknown type for if condition: %v", condition.T))
      }
    }
    case "defn": {
      if expanded[1].T != TypeID { panic(fmt.Sprintf("Expected ID, got: %v", expanded[1].T)) }
      if expanded[2].T != TypeExpression { panic("second function arguments must be a list") }

      fnName, _ := expanded[1].V.(string)
      fnArgs, _ := expanded[2].V.(*Stack)

      fun := CreateFunction(context, fnArgs, expanded[3])

      context.vars[fnName] = fun
      return fun
    }
    default: panic(fmt.Sprintf("unknown f: %v", fnName))
    }
  }
  case TypeFunction: {
    fv, _ := f.V.(ValueFunction)
    argsCount := uint32(len(fv.argsNames))
    // argsNames, boundContext, body
    newContext := EmptyEvaluationContext()//
    for k, v := range fv.boundContext.vars {
      newContext.vars[k] = v
    }

    if (argsCount != s.Size - 1) { panic("argument count missmatch") }

    for i, v := range expanded[1:] {
      xv := v.Evaluate(context)
      newContext.vars[fv.argsNames[i]] = xv
    }

    return fv.body.Evaluate(newContext)
  }
  default: panic("Unknown type")
  }

  panic("some shit happened")
}
func (v *Value) Evaluate(context *EvaluationContext) *Value {
  switch v.T {
  case TypeID: {
    key, _ := v.V.(string)
    val, ok := context.vars[key]

    if ok { return val }
          { return v }
  }
  case TypeNumber: return v
  case TypeExpression: {
    s, _ := v.V.(*Stack)
    return s.Evaluate(context)
  }
  default: panic("Unknown type")
  }
}
