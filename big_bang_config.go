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
