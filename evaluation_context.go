package main

type EvaluationContext struct {
  vars map[string]*Value
}
func EmptyEvaluationContext() *EvaluationContext {
  return &EvaluationContext{vars: make(map[string]*Value)}
}

func (c *EvaluationContext) Clone() *EvaluationContext {
  newContext := EmptyEvaluationContext()
  for key, value := range c.vars {
    newContext.vars[key] = value
  }
  return newContext
}
