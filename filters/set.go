package filters

import proto "github.com/codeuniversity/al-proto"

//Set of Filters
type Set []*Filter

//SetFromDefinitions ...
func SetFromDefinitions(definitions []*FilterDefinition) Set {
	set := Set{}
	for _, definition := range definitions {
		set = append(set, NewFilter(definition))
	}
	return set
}

//Eval all filters to see if the cell passes all of them
func (s Set) Eval(cell *proto.Cell) (passes bool, warnings []string) {
	passes = true
	for _, filter := range s {
		filterPassed, filterWarnings := filter.Eval(cell)
		if len(filterWarnings) > 0 {
			warnings = append(warnings, filterWarnings...)
		}
		if !filterPassed {
			passes = false
		}
	}

	return
}
