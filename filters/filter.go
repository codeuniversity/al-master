package filters

import (
	"fmt"

	"github.com/codeuniversity/al-proto"
)

//FilterDefinition ...
type FilterDefinition struct {
	LeftHand      string `json:"left_hand"`
	LeftHandType  string `json:"left_hand_type"`
	Operator      string `json:"operator"`
	RightHand     string `json:"right_hand"`
	RightHandType string `json:"right_hand_type"`
}

type operator int

const (
	operatorInvalid     operator = iota
	operatorLessThan    operator = iota
	operatorGreaterThan operator = iota
	operatorEqual       operator = iota
)

//Filter ...
type Filter struct {
	leftVar Var
	rgtVar  Var
	op      operator
}

//NewFilter from FilterDefinition. Compiles the two sides as far as it can beforehand
func NewFilter(definition *FilterDefinition) *Filter {
	var op operator
	switch definition.Operator {
	case "<":
		op = operatorLessThan
	case ">":
		op = operatorGreaterThan
	case "=":
		op = operatorEqual
	default:
		op = operatorInvalid
	}

	return &Filter{
		leftVar: compileHand(definition.LeftHand, definition.LeftHandType),
		rgtVar:  compileHand(definition.RightHand, definition.RightHandType),
		op:      op,
	}
}

//Eval filter to see if the cell passes this Filter
func (f *Filter) Eval(cell *proto.Cell) (passes bool, warnings []string) {
	lftVar := f.leftVar.Eval(cell)
	if !lftVar.Valid() {
		warnings = append(warnings, fmt.Sprintf("left hand %v is invalid", *lftVar.(*invalidVar)))
	}
	rgtVar := f.rgtVar.Eval(cell)
	if !rgtVar.Valid() {
		warnings = append(warnings, fmt.Sprintf("right hand %v is invalid", *rgtVar.(*invalidVar)))
	}

	switch f.op {
	case operatorLessThan:
		passes = lftVar.LessThan(rgtVar)
	case operatorGreaterThan:
		passes = lftVar.GreaterThan(rgtVar)
	case operatorEqual:
		passes = lftVar.Equal(rgtVar)
	case operatorInvalid:
		warnings = append(warnings, "operator is invalid")
	}

	return
}

func compileHand(v, t string) Var {
	switch t {
	case "number":
		return compileNumberVar(v)
	case "coordinate":
		return compileCoordinateVar(v)
	}

	return &invalidVar{Value: v, Type: t}
}
