package main

type EvaluationContext struct {
  vars map[string]*Value
}
func EmptyEvaluationContext() *EvaluationContext {
  return &EvaluationContext{vars: make(map[string]*Value)}
}

