package filters

import (
	"strconv"

	"github.com/codeuniversity/al-proto"
)

//Var is one side of a filter
type Var interface {
	Eval(*proto.Cell) Var
	Valid() bool

	LessThan(Var) bool
	GreaterThan(Var) bool
	Equal(Var) bool
}

type numberVar struct {
	Value float32
}

func compileNumberVar(rawValue string) Var {
	value, err := strconv.ParseFloat(rawValue, 32)
	if err != nil {
		return &invalidVar{Value: rawValue, Type: "number"}
	}

	return &numberVar{Value: float32(value)}
}

func (v *numberVar) Eval(*proto.Cell) Var {
	return v
}
func (v *numberVar) Valid() bool {
	return true
}

func (v *numberVar) LessThan(other Var) bool {
	switch other.(type) {
	case *numberVar:
		return v.Value < (other.(*numberVar)).Value
	}

	return false
}

func (v *numberVar) GreaterThan(other Var) bool {
	switch other.(type) {
	case *numberVar:
		return v.Value > (other.(*numberVar)).Value
	}

	return false
}
func (v *numberVar) Equal(other Var) bool {
	switch other.(type) {
	case *numberVar:
		return v.Value == (other.(*numberVar)).Value
	}

	return false

}

type coordinate int

const (
	coordinateX coordinate = iota
	coordinateY coordinate = iota
	coordinateZ coordinate = iota
)

type coordinateVar struct {
	Coordinate coordinate
}

func compileCoordinateVar(rawValue string) Var {
	switch rawValue {
	case "cell.pos.x":
		return &coordinateVar{Coordinate: coordinateX}
	case "cell.pos.y":
		return &coordinateVar{Coordinate: coordinateY}
	case "cell.pos.z":
		return &coordinateVar{Coordinate: coordinateZ}
	}

	return &invalidVar{Value: rawValue, Type: "coordinate"}
}

func (v *coordinateVar) Eval(cell *proto.Cell) Var {
	switch v.Coordinate {
	case coordinateX:
		return &numberVar{Value: cell.Pos.X}
	case coordinateY:
		return &numberVar{Value: cell.Pos.Y}
	case coordinateZ:
		return &numberVar{Value: cell.Pos.Z}
	}

	return &invalidVar{}
}

func (v *coordinateVar) Valid() bool {
	return true
}

func (v *coordinateVar) LessThan(other Var) bool {
	return false
}
func (v *coordinateVar) GreaterThan(other Var) bool {
	return false
}
func (v *coordinateVar) Equal(other Var) bool {
	return false
}

type invalidVar struct {
	Value string
	Type  string
}

func (v *invalidVar) Eval(*proto.Cell) Var {
	return v
}

func (v *invalidVar) Valid() bool {
	return false
}

func (v *invalidVar) LessThan(other Var) bool {
	return false
}
func (v *invalidVar) GreaterThan(other Var) bool {
	return false
}
func (v *invalidVar) Equal(other Var) bool {
	return false
}
