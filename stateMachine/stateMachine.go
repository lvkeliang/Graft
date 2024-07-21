package stateMachine

type StateMachine interface {
	ToApply(string) string
}
