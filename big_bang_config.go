package master

import (
	"io/ioutil"

	"github.com/codeuniversity/al-proto"
	"gopkg.in/yaml.v2"
)

//BigBangConfig is the structure of the provided  big_bang_config.yaml
type BigBangConfig struct {
	SpawnDimension `yaml:"spawn_dimension"`
	EnergyLevel    uint64 `yaml:"energy_level"`
	CellAmount     uint64 `yaml:"cell_amount"`
	DnaLengthRange `yaml:"dna_length_range"`
}

//SpawnDimension ...
type SpawnDimension struct {
	Start Vector `yaml:"start"`
	End   Vector `yaml:"end"`
}

// DnaLengthRange ...
type DnaLengthRange struct {
	Min uint64 `yaml:"min"`
	Max uint64 `yaml:"max"`
}

//Vector ...
type Vector struct {
	X float32 `yaml:"x"`
	Y float32 `yaml:"y"`
	Z float32 `yaml:"z"`
}

//BigBangConfigFromPath loads the yaml-file the path is pointing to and returns a BigBangConfig with values on success.
//returns an error if it couldn't load the file or unmarshal the content
func BigBangConfigFromPath(path string) (*BigBangConfig, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &BigBangConfig{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

//ToProto from the yaml friendly vector type
func (v *Vector) ToProto() *proto.Vector {
	return &proto.Vector{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}
}

//ToProto from the yaml friendly SpawnDimension type
func (d *SpawnDimension) ToProto() *proto.SpawnDimension {
	return &proto.SpawnDimension{
		Start: d.Start.ToProto(),
		End:   d.End.ToProto(),
	}
}

//ToProto from the yaml friendly DnaLengthRange type
func (r *DnaLengthRange) ToProto() *proto.DnaLengthRange {
	return &proto.DnaLengthRange{
		Min: r.Min,
		Max: r.Max,
	}
}

//ToProto from the yaml friendly BigBangConfig type
func (c *BigBangConfig) ToProto() *proto.BigBangRequest {
	return &proto.BigBangRequest{
		SpawnDimension: c.SpawnDimension.ToProto(),
		EnergyLevel:    c.EnergyLevel,
		CellAmount:     c.CellAmount,
		DnaLengthRange: c.DnaLengthRange.ToProto(),
	}
}
