package yasp

import (
  "strconv"
  "container/list"
  "bytes"
)

type Parsing struct {
  stack *Stack
  expressions *list.List

  parsingString *bytes.Buffer
}

func (p *Parsing) Init() {
  p.stack = CreateStack(nil)
  p.parsingString = new(bytes.Buffer)
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

func (p *Parsing) StartString() {
  p.parsingString.Reset()
}
func (p *Parsing) EndString() {
  p.stack.AddToStack(Value{T: TypeString, V: p.parsingString.String()})
  p.parsingString = new(bytes.Buffer)
}

func (p *Parsing) AddCharacter(str string) {
  p.parsingString.WriteString(str)
}
