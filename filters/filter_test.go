package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codeuniversity/al-proto"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	t.Run("filter with number vars on both sides works", func(t *testing.T) {
		definition := &FilterDefinition{
			LeftHand:      "30",
			LeftHandType:  "number",
			Operator:      "<",
			RightHand:     "42",
			RightHandType: "number",
		}

		filter := NewFilter(definition)
		require.NotNil(t, filter)
		someCell := &proto.Cell{}

		passed, warnings := filter.Eval(someCell)
		assert.True(t, passed)
		assert.Empty(t, warnings)
	})

	t.Run("filter with coordinate var on one side works", func(t *testing.T) {
		definition := &FilterDefinition{
			LeftHand:      "cell.pos.x",
			LeftHandType:  "coordinate",
			Operator:      "<",
			RightHand:     "42",
			RightHandType: "number",
		}

		filter := NewFilter(definition)
		require.NotNil(t, filter)
		cell := &proto.Cell{Pos: &proto.Vector{X: 1, Y: 2, Z: 3}}

		passed, warnings := filter.Eval(cell)
		assert.True(t, passed)
		assert.Empty(t, warnings)
	})

	t.Run("filter with coordinate vars on both sides works", func(t *testing.T) {
		definition := &FilterDefinition{
			LeftHand:      "cell.pos.x",
			LeftHandType:  "coordinate",
			Operator:      "<",
			RightHand:     "cell.pos.y",
			RightHandType: "coordinate",
		}

		filter := NewFilter(definition)
		require.NotNil(t, filter)
		cell := &proto.Cell{Pos: &proto.Vector{X: 1, Y: 2, Z: 3}}

		passed, warnings := filter.Eval(cell)
		assert.True(t, passed)
		assert.Empty(t, warnings)
	})

	t.Run("filter with incorrect definition outputs warnings", func(t *testing.T) {
		definition := &FilterDefinition{
			LeftHand:      "celll.pos.x",
			LeftHandType:  "cordinate",
			Operator:      "foo",
			RightHand:     "cell.pos.y",
			RightHandType: "unmber",
		}

		filter := NewFilter(definition)
		require.NotNil(t, filter)
		cell := &proto.Cell{Pos: &proto.Vector{X: 1, Y: 2, Z: 3}}

		passed, warnings := filter.Eval(cell)
		assert.False(t, passed)
		assert.Equal(
			t,
			[]string{
				"left hand {celll.pos.x cordinate} is invalid",
				"right hand {cell.pos.y unmber} is invalid",
				"operator is invalid",
			},
			warnings,
		)
	})
}
