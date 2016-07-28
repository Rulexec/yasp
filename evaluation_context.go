package yasp

type EvaluationContext struct {
  Vars map[string]*Value
}
func EmptyEvaluationContext() *EvaluationContext {
  return &EvaluationContext{Vars: make(map[string]*Value)}
}

func (c *EvaluationContext) Clone() *EvaluationContext {
  newContext := EmptyEvaluationContext()
  for key, value := range c.Vars {
    newContext.Vars[key] = value
  }
  return newContext
}
