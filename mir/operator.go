package mir

type OperatorType uint

const (
	OperatorAdd OperatorType = iota
	OperatorSub
	OperatorMul
	OperatorDiv
	OperatorEq
	OperatorLt
	OperatorGt
	OperatorNe
	OperatorLe
	OperatorGe
	OperatorAnd
	OperatorOr
	OperatorPow
)
