package filters

import (
	"testing"

	proto "github.com/codeuniversity/al-proto"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	t.Run("Set checks all Filters to see if a cell should pass", func(t *testing.T) {
		passingCell := &proto.Cell{Pos: &proto.Vector{X: 1, Y: 2, Z: 3}}
		notPassingCell := &proto.Cell{Pos: &proto.Vector{X: 100, Y: 200, Z: 300}}

		definitions := []*FilterDefinition{
			&FilterDefinition{
				LeftHand:      "cell.pos.x",
				LeftHandType:  "coordinate",
				Operator:      "<",
				RightHand:     "42",
				RightHandType: "number",
			},
			&FilterDefinition{
				LeftHand:      "cell.pos.y",
				LeftHandType:  "coordinate",
				Operator:      "<",
				RightHand:     "100",
				RightHandType: "number",
			},
		}
		set := SetFromDefinitions(definitions)

		passed, warnings := set.Eval(passingCell)
		assert.True(t, passed)
		assert.Empty(t, warnings)

		passed, warnings = set.Eval(notPassingCell)
		assert.False(t, passed)
		assert.Empty(t, warnings)
	})

	t.Run("concats warnings from filters", func(t *testing.T) {
		someCell := &proto.Cell{Pos: &proto.Vector{X: 1, Y: 2, Z: 3}}

		definitions := []*FilterDefinition{
			&FilterDefinition{
				LeftHand:      "cell.pos.x",
				LeftHandType:  "coordinate",
				Operator:      "foo",
				RightHand:     "42",
				RightHandType: "number",
			},
			&FilterDefinition{
				LeftHand:      "cell.pos.y",
				LeftHandType:  "coordinate",
				Operator:      "<",
				RightHand:     "100",
				RightHandType: "nummber",
			},
		}
		set := SetFromDefinitions(definitions)
		passed, warnings := set.Eval(someCell)
		assert.False(t, passed)
		assert.Equal(
			t,
			[]string{
				"operator is invalid",
				"right hand {100 nummber} is invalid",
			},
			warnings)
	})
}
