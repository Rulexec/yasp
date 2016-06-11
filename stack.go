package main

import ( "bytes" )

type StackNode struct {
  Prev *StackNode
  V Value
}
type Stack struct {
  Prev *Stack

  Top *StackNode
  Size uint32
}

func (s *Stack) String() string {
  var buffer bytes.Buffer

  var f func (node *StackNode)
  f = func (node *StackNode) {
    if (node == nil) { return }

    f(node.Prev)

    buffer.WriteString(node.V.String())
    buffer.WriteString(", ")
  }

  if (s.Top != nil) {
    buffer.WriteString("[ ")

    f(s.Top.Prev);
    buffer.WriteString(s.Top.V.String());

    buffer.WriteString(" ]")
  } else {
    buffer.WriteString("[]")
  }

  return buffer.String()
}

func CreateStack(prev *Stack) *Stack {
  return &Stack{Prev: prev, Top: nil, Size: 0}
}

func (s *Stack) AddToStack(value Value) {
  if (s.Top == nil) {
    s.Top = &StackNode{Prev: nil, V: value}
  } else {
    newNode := &StackNode{Prev: s.Top, V: value}
    s.Top = newNode
  }

  s.Size++
}
func (s *Stack) IsEmpty() bool {
  return s.Top == nil
}

func (s *Stack) Expand() []*Value {
  expanded := make([]*Value, s.Size)
  for cur, i := s.Top, s.Size - 1 ;
      cur != nil;
      cur, i = cur.Prev, i - 1 {
    expanded[i] = &cur.V;
  }

  return expanded;
}

