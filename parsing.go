package main

import (
  "strconv"
  "container/list"
)

type Parsing struct {
  stack *Stack
  expressions *list.List
}

func (p *Parsing) Init() {
  p.stack = CreateStack(nil)
}

func (p *Parsing) OpenBrace() {
  newStack := CreateStack(p.stack)
  p.stack = newStack
}
func (p *Parsing) CloseBrace() {
  if (p.stack.Prev == nil) { return }

  cur := p.stack
  p.stack = p.stack.Prev
  p.stack.AddToStack(Value{T: TypeExpression, V: cur})
}

func (p *Parsing) AddID(id string) {
  p.stack.AddToStack(Value{T: TypeID, V: id})
}
func (p *Parsing) AddNumber(number string) {
  n, err := strconv.ParseUint(number, 10, 32)
  if (err != nil) { panic(err) }
  p.stack.AddToStack(Value{T: TypeNumber, V: uint64(n)})
}
