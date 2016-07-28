package yasp

import ( "fmt" )

func (v *Value) AssertIdType() string {
  if v.T != TypeID { panic(fmt.Sprintf("%v expected to be ID", v)) }

  s, _ := v.V.(string)

  return s
}
func (v *Value) AssertNumberType() uint64 {
  if v.T != TypeNumber { panic(fmt.Sprintf("%v expected to be Number", v)) }

  n, _ := v.V.(uint64)

  return n
}
func (v *Value) AssertFunctionType() *ValueFunction {
  if v.T != TypeFunction { panic(fmt.Sprintf("%v expected to be Function", v)) }

  vf, _ := v.V.(ValueFunction)

  return &vf
}
func (v *Value) AssertStringType() string {
  if v.T != TypeString { panic(fmt.Sprintf("%v expected to be String", v)) }

  s, _ := v.V.(string)

  return s
}
func (v *Value) AssertExpressionType() *Stack {
  if v.T != TypeExpression { panic(fmt.Sprintf("%v expected to be Expression", v)) }

  s, _ := v.V.(*Stack)

  return s
}
func (v *Value) AssertListType() []*Value {
  if v.T != TypeList { panic(fmt.Sprintf("%v expected to be List", v)) }

  l, _ := v.V.([]*Value)

  return l
}
