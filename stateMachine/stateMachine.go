package stateMachine

type StateMachine interface {
	ToApply(string interface{}) interface{}
}
