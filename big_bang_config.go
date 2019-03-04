package master
import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/codeuniversity/al-proto"
)


type BigBangConfig struct {
	SpawnDimension	`yaml:"spawn_dimension"`
	EnergyLevel uint64 `yaml:"energy_level"`
	CellAmount uint64	`yaml:"cell_amount"`
	DnaLengthRange	`yaml:"dna_length_range"`
}

type SpawnDimension struct {
	Start Vector `yaml:"start"`
	End Vector `yaml:"end"`
}

type DnaLengthRange struct {
	Min uint64 `yaml:"min"`
	Max uint64 `yaml:"max"`
}

type Vector struct {
	X float32 `yaml:"x"`
	Y float32 `yaml:"y"`
	Z float32 `yaml:"z"`
}

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

func (v *Vector) ToProto() *proto.Vector {
	return &proto.Vector{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}
}

func (d *SpawnDimension) ToProto() *proto.SpawnDimension {
	return &proto.SpawnDimension{
		Start: d.Start.ToProto(),
		End: d.End.ToProto(),
	}
}

func (r *DnaLengthRange) ToProto() *proto.DnaLengthRange {
	return &proto.DnaLengthRange{
		Min: r.Min,
		Max: r.Max,
	}
}

func (c *BigBangConfig) ToProto() *proto.BigBangRequest{
	return &proto.BigBangRequest{
		SpawnDimension: c.SpawnDimension.ToProto(),
		EnergyLevel: c.EnergyLevel,
		CellAmount: c.CellAmount,
		DnaLengthRange: c.DnaLengthRange.ToProto(),
	}
}
