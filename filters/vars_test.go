package filters

import (
	"testing"

	"github.com/codeuniversity/al-proto"

	"github.com/stretchr/testify/assert"
)

func TestVars(t *testing.T) {
	t.Run("number vars work as intended", func(t *testing.T) {
		lftVar := &numberVar{Value: 1}
		rgtVar := &numberVar{Value: 2}

		assert.True(t, lftVar.LessThan(rgtVar))
		assert.False(t, lftVar.GreaterThan(rgtVar))
		assert.False(t, lftVar.Equal(rgtVar))
		assert.True(t, lftVar.Equal(lftVar))
	})

	t.Run("coordinate vars eval into correct number vars", func(t *testing.T) {
		coordinateXVar := &coordinateVar{Coordinate: coordinateX}
		coordinateYVar := &coordinateVar{Coordinate: coordinateY}
		coordinateZVar := &coordinateVar{Coordinate: coordinateZ}
		numberVar := &numberVar{Value: 4}
		firstCell := &proto.Cell{Pos: &proto.Vector{X: 1, Y: 2, Z: 3}}

		assert.True(t, coordinateXVar.Eval(firstCell).LessThan(numberVar))
		assert.True(t, coordinateYVar.Eval(firstCell).LessThan(numberVar))
		assert.True(t, coordinateZVar.Eval(firstCell).LessThan(numberVar))
		assert.False(t, coordinateXVar.Eval(firstCell).GreaterThan(numberVar))
		assert.False(t, coordinateYVar.Eval(firstCell).GreaterThan(numberVar))
		assert.False(t, coordinateZVar.Eval(firstCell).GreaterThan(numberVar))

		secondCell := &proto.Cell{Pos: &proto.Vector{X: 4, Y: 5, Z: 6}}

		assert.True(t, coordinateXVar.Eval(secondCell).Equal(numberVar))
		assert.True(t, coordinateYVar.Eval(secondCell).GreaterThan(numberVar))
		assert.True(t, coordinateZVar.Eval(secondCell).GreaterThan(numberVar))
		assert.False(t, coordinateYVar.Eval(secondCell).LessThan(numberVar))
		assert.False(t, coordinateZVar.Eval(secondCell).LessThan(numberVar))
	})
}
