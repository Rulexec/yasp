package yasp

import (
  "fmt"
  "bytes"
)

func (p *Parsing) Evaluate(context *EvaluationContext) *Value {
  if (p.stack.IsEmpty()) {
    return nil
  }

  expanded := p.stack.Expand()

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
func (s *Stack) Evaluate(context *EvaluationContext) *Value {
  if s.Size == 0 { return &Value{T: TypeNil} }

  expanded := s.Expand()

  f := expanded[0].Evaluate(context)

  if (s.Prev == nil && s.Size == 1) { return f }

  switch f.T {
  case TypeNumber: fmt.Printf("1 %v ", s); panic("Number is not a function")
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
    case "*": {
     var result uint64 = 1

     for _, x := range expanded[1:] {
       xv := x.Evaluate(context)
       if (xv.T != TypeNumber) { panic("is not a number for *") }

       n, _ := xv.V.(uint64)
       result *= n
     }

     return &Value{T: TypeNumber, V: result}
    }
    case "listToString": {
      AssertNumberOfArguments(s, 1, fnName)

      var buffer bytes.Buffer

      lst := expanded[1].Evaluate(context).AssertListType()

      for _, x := range lst {
        s := x.Evaluate(context).AssertStringType()
        buffer.WriteString(s)
      }

      return &Value{T: TypeString, V: buffer.String()}
    }
    case "or": {
     for _, x := range expanded[1:] {
       xv := x.Evaluate(context)
       if xv.Bool() { return &Value{T: TypeNumber, V: uint64(1)} }
     }

     return &Value{T: TypeNumber, V: uint64(0)}
    }
    case "and": {
     for _, x := range expanded[1:] {
       xv := x.Evaluate(context)
       if !xv.Bool() { return &Value{T: TypeNumber, V: uint64(0)} }
     }

     return &Value{T: TypeNumber, V: uint64(1)}
    }
    case "ord": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context).AssertStringType()
      c := uint64([]rune(x)[0])

      return &Value{T: TypeNumber, V: c}
    }
    case "-": {
      AssertNumberOfArguments(s, 2, fnName)

      a, b := expanded[1].Evaluate(context).AssertNumberType(),
              expanded[2].Evaluate(context).AssertNumberType()

      return &Value{T: TypeNumber, V: a - b}
    }
    case "=": {
      AssertNumberOfArguments(s, 2, fnName)

      a, b := expanded[1].Evaluate(context), expanded[2].Evaluate(context)

      if a.Equals(b) {
        return &Value{T: TypeNumber, V: uint64(1)}
      } else {
        return &Value{T: TypeNumber, V: uint64(0)}
      }
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
    case ">=": {
      AssertNumberOfArguments(s, 2, fnName)

      a, b := expanded[1].Evaluate(context).AssertNumberType(),
              expanded[2].Evaluate(context).AssertNumberType()

      if a >= b {
        return &Value{T: TypeNumber, V: uint64(1)}
      } else {
        return &Value{T: TypeNumber, V: uint64(0)}
      }
    }
    case "<=": {
      AssertNumberOfArguments(s, 2, fnName)

      a, b := expanded[1].Evaluate(context).AssertNumberType(),
              expanded[2].Evaluate(context).AssertNumberType()

      if a <= b {
        return &Value{T: TypeNumber, V: uint64(1)}
      } else {
        return &Value{T: TypeNumber, V: uint64(0)}
      }
    }
    case "let": {
      AssertNumberOfArguments(s, 2, fnName)

      stack := expanded[1].AssertExpressionType()
      expandedLet := stack.Expand()

      letCount := len(expandedLet)

      if letCount % 2 != 0 { panic("let must have even number of values") }
      letCount /= 2

      newContext := context.Clone()

      for i := 0; i < letCount; i++ {
        varName := expandedLet[2 * i].AssertIdType()
        varValue := expandedLet[2 * i + 1].Evaluate(newContext)

        newContext.Vars[varName] = varValue
      }

      return expanded[2].Evaluate(newContext)
    }
    case "fn": {
      if s.Size != 3 { fmt.Printf("2 %v ", s); panic("function must have 2 args") }

      if expanded[1].T != TypeExpression { panic("second function arguments must be a list") }

      argsStack, _ := expanded[1].V.(*Stack)
      return CreateFunction(context, argsStack, expanded[2])
    }
    case "head": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeString: return &Value{T: TypeString, V: string([]rune(x.AssertStringType())[0])}
      case TypeList: {
        lst := x.AssertListType()
        return lst[0]
      }
      default: fmt.Printf("%v\n", x); panic(fmt.Sprintf("Unknown type for head: %v", x.T))
      }
    }
    case "take": {
      AssertNumberOfArguments(s, 2, fnName)

      n := expanded[1].Evaluate(context).AssertNumberType()
      collection := expanded[2].Evaluate(context)

      switch collection.T {
      case TypeString: {
        s, _ := collection.V.(string)
        runes := []rune(s)
        taken := string(runes[:n])

        return &Value{T: TypeString, V: taken}
      }
      case TypeList: {
        lst := collection.AssertListType()
        return &Value{T: TypeList, V: lst[:n]}
      }
      default: panic(fmt.Sprintf("Unknown type for take: %v", collection.T))
      }
    }
    case "skip": {
      AssertNumberOfArguments(s, 2, fnName)

      n := expanded[1].Evaluate(context).AssertNumberType()
      collection := expanded[2].Evaluate(context)

      switch collection.T {
      case TypeString: {
        s, _ := collection.V.(string)
        runes := []rune(s)
        skipped := string(runes[n:])

        return &Value{T: TypeString, V: skipped}
      }
      case TypeList: {
        lst := collection.AssertListType()
        return &Value{T: TypeList, V: lst[n:]}
      }
      default: panic(fmt.Sprintf("Unknown type for skip: %v", collection.T))
      }
    }
    case "not": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeNumber: {
        n, _ := x.V.(uint64)
        if n == 0 { return &Value{T: TypeNumber, V: uint64(1)}
        } else { return &Value{T: TypeNumber, V: uint64(0)} }
      }
      default: fmt.Printf("%v\n", x); panic(fmt.Sprintf("Unknown type for not: %v", x.T))
      }
    }
    case "len": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeString: {
        s, _ := x.V.(string)
        return &Value{T: TypeNumber, V: uint64(len([]rune(s)))}
      }
      default: fmt.Printf("%v\n", x); panic(fmt.Sprintf("Unknown type for len: %v", x.T))
      }
    }
    case "last": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeList: {
        lst := x.AssertListType()
        return lst[len(lst) - 1]
      }
      default: fmt.Printf("%v\n", x); panic(fmt.Sprintf("Unknown type for last: %v", x.T))
      }
    }
    case "tail": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeString: {
        s, _ := x.V.(string)
        runes := []rune(s)
        tail := string(runes[1:])

        return &Value{T: TypeString, V: tail}
      }
      case TypeList: {
        lst, _ := x.V.([]*Value)
        return &Value{T: TypeList, V: lst[1:]}
      }
      default: panic(fmt.Sprintf("Unknown type for tail: %v", x.T))
      }
    }
    case "untail": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeList: {
        lst := x.AssertListType()
        return &Value{T: TypeList, V: lst[:len(lst) - 1]}
      }
      default: panic(fmt.Sprintf("Unknown type for untail: %v", x.T))
      }
    }
    case "typeof": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeID: return &Value{T: TypeString, V: "id"}
      case TypeNumber: return &Value{T: TypeString, V: "number"}
      case TypeString: return &Value{T: TypeString, V: "string"}
      case TypeExpression: return &Value{T: TypeString, V: "expression"}
      case TypeFunction: return &Value{T: TypeString, V: "function"}
      case TypeNil: return &Value{T: TypeString, V: "nil"}
      case TypeList: return &Value{T: TypeString, V: "list"}
      default: panic(fmt.Sprintf("Unknown type for typeof: %v", x.T))
      }
    }
    case "empty?": {
      AssertNumberOfArguments(s, 1, fnName)

      x := expanded[1].Evaluate(context)

      switch x.T {
      case TypeString: {
        s, _ := x.V.(string)
        if s == "" { return &Value{T: TypeNumber, V: uint64(1)}
        } else { return &Value{T: TypeNumber, V: uint64(0)} }
      }
      case TypeList: {
        lst, _ := x.V.([]*Value)
        if len(lst) == 0 {
          return &Value{T: TypeNumber, V: uint64(1)}
        } else {
          return &Value{T: TypeNumber, V: uint64(0)}
        }
      }
      default: panic(fmt.Sprintf("Unknown type for empty?: %v", x.T))
      }
    }
    case "in": {
      AssertNumberOfArguments(s, 2, fnName)

      key := expanded[1].Evaluate(context)
      collection := expanded[2].Evaluate(context)

      switch collection.T {
      case TypeList: {
        lst := collection.AssertListType()
        for _, x := range lst {
          if key.Equals(x) { return &Value{T: TypeNumber, V: uint64(1)} }
        }
        return &Value{T: TypeNumber, V: uint64(0)}
      }
      default: panic(fmt.Sprintf("Unknown type for in: %v", collection.T))
      }
    }
    case "get": {
      AssertNumberOfArguments(s, 2, fnName)

      i := expanded[1].Evaluate(context).AssertNumberType()
      x := expanded[2].Evaluate(context)

      switch x.T {
      case TypeList: {
        lst := x.AssertListType()
        return lst[i]
      }
      default: fmt.Printf("%v\n", expanded); panic(fmt.Sprintf("Unknown type for get: %v", x.T))
      }
    }
    case "getOrDef": {
      AssertNumberOfArguments(s, 3, fnName)

      def := expanded[1]
      i := expanded[2].Evaluate(context).AssertNumberType()
      x := expanded[3].Evaluate(context)

      switch x.T {
      case TypeList: {
        lst := x.AssertListType()
        if i < uint64(len(lst)) { return lst[i]
        } else { def.Evaluate(context) }
      }
      default: panic(fmt.Sprintf("Unknown type for getOrDef: %v", x.T))
      }
    }
    case "do": {
      var last *Value

      newContext := context.Clone()

      for _, x := range expanded[1:] {
        last = x.Evaluate(newContext)
      }

      return last
    }
    case "switch": {
      xv := expanded[1].Evaluate(context)

      odd := s.Size % 2

      for i := uint32(0); i < (s.Size - 2) / 2 - odd; i++ {
        key := expanded[2 + 2 * i]

        if key.Equals(xv) {
          return expanded[2 + 2 * i + 1].Evaluate(context)
        }
      }

      if odd == 1 {
        return expanded[s.Size - 1].Evaluate(context)
      } else {
        fmt.Printf("%v\n%v\n", expanded, xv)
        panic("No default branch in switch")
      }
    }
    case "concat": {
      var buffer bytes.Buffer

      for _, x := range expanded[1:] {
        xv := x.Evaluate(context)
        s := xv.AssertStringType()

        buffer.WriteString(s)
      }

      return &Value{T: TypeString, V: buffer.String()}
    }
    case "append": {
      AssertNumberOfArguments(s, 2, fnName)

      x := expanded[1].Evaluate(context)
      collection := expanded[2].Evaluate(context)

      switch collection.T {
      case TypeList: {
        lst := collection.AssertListType()
        return &Value{T: TypeList, V: append(lst, x)}
      }
      default: fmt.Printf("%v\n", expanded); panic(fmt.Sprintf("Unknown type for append: %v", collection.T))
      }
    }
    case "prepend": {
      AssertNumberOfArguments(s, 2, fnName)

      x := expanded[1].Evaluate(context)
      collection := expanded[2].Evaluate(context)

      switch collection.T {
      case TypeList: {
        lst := collection.AssertListType()
        return &Value{T: TypeList, V: append([]*Value{x}, lst...)}
      }
      default: fmt.Printf("%v\n", expanded); panic(fmt.Sprintf("Unknown type for prepend: %v", collection.T))
      }
    }
    case "print": {
      first := expanded[1].Evaluate(context)
      fmt.Println(first.V)

      for _, x := range expanded[2:] {
        xv := x.Evaluate(context)
        fmt.Println(xv.V)
      }

      return first
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
      default: fmt.Printf("%v\n", expanded); panic(fmt.Sprintf("Unknown type for if condition: %v", condition.T))
      }
    }
    case "def": {
      key := expanded[1].AssertIdType()
      context.Vars[key] = expanded[2].Evaluate(context)

      return expanded[2]
    }
    case "defn": {
      if expanded[1].T != TypeID { panic(fmt.Sprintf("Expected ID, got: %v", expanded[1].T)) }
      if expanded[2].T != TypeExpression { panic("second function arguments must be a list") }

      fnName, _ := expanded[1].V.(string)
      fnArgs, _ := expanded[2].V.(*Stack)

      fun := CreateFunction(context, fnArgs, expanded[3])

      context.Vars[fnName] = fun
      return fun
    }
    case "list": {
      var lst []*Value

      for _, x := range expanded[1:] {
        xv := x.Evaluate(context)
        lst = append(lst, xv)
      }

      return &Value{T: TypeList, V: lst}
    }
    default: panic(fmt.Sprintf("unknown f: %v", fnName))
    }
  }
  case TypeFunction: { return f.EvaluateFunction(context, expanded[1:]) }
  default: fmt.Printf("x: %v\n", expanded); panic(fmt.Sprintf("Unknown type: %v", f.T))
  }

  panic("some shit happened")
}
func (v *Value) Evaluate(context *EvaluationContext) *Value {
  switch v.T {
  case TypeID: {
    key, _ := v.V.(string)
    val, ok := context.Vars[key]

    if ok { return val }
          { return v }
  }
  case TypeNumber, TypeString: return v
  case TypeExpression: {
    s, _ := v.V.(*Stack)
    return s.Evaluate(context)
  }
  default: panic(fmt.Sprintf("Unknown type: %v", v.T))
  }
}

func (v *Value) EvaluateFunction(context *EvaluationContext, args []*Value) *Value {
  fv := v.AssertFunctionType()
  argsCount := uint32(len(args))
  newContext := EmptyEvaluationContext()
  for k, v := range fv.boundContext.Vars {
    newContext.Vars[k] = v
  }

  if (argsCount != uint32(len(fv.argsNames))) { panic(fmt.Sprintf("argument count missmatch %v", args)) }

  for i, v := range args {
    xv := v.Evaluate(context)
    newContext.Vars[fv.argsNames[i]] = xv
  }

  return fv.body.Evaluate(newContext)
}
