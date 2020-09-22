package patch

import (
	"reflect"
	"strings"
)

var mapper *Mapper

// init mapper when the package is imported
func init() {
	if mapper == nil {
		mapper = &Mapper{
			inputType:  "json",
			outputType: "boil",
			modelMaps:  make(map[string]Map),
		}
	}
}

// Mapper is a component loading and saving in memory patch.Maps for requested models in real time
// everytime a patch.Maps is requested for a model, the mapper is involved in order to know
// either the model has been already mapped or not
// this allows us to mitigate the use of struct reflection
type Mapper struct {
	// configure the kind of tag the mapper is using for input & output
	// examples: `json`, `xml`, `gorm`, `boil`
	inputType  string
	outputType string

	// model PatchMaps memory storage
	modelMaps map[string]Map
}

func (m *Mapper) loadOrInitMap(model interface{}) Map {
	// get name from model
	t := reflect.TypeOf(model)
	name := ""
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name = t.Name()

	// try to load map
	modelMap, ok := mapper.modelMaps[name]
	if !ok {
		// load map if not existing yet
		return mapper.initMap(t, name)
	}
	return modelMap
}

func (m *Mapper) initMap(model reflect.Type, name string) Map {
	modelMap := Map{
		input:  make(map[string]string),
		output: make(map[string]string),
	}

	// parse fields and look input and output tags up
	n := model.NumField()
	for i := 0; i < n; i++ {
		field := model.Field(i)
		// prepare input match
		inputV, ok := m.getTagValue(m.inputType, field)
		if ok {
			modelMap.input[inputV] = field.Name
		}
		outputV, ok := m.getTagValue(m.outputType, field)
		if ok {
			modelMap.output[field.Name] = outputV
		}
	}
	mapper.modelMaps[name] = modelMap
	return modelMap
}

func (m *Mapper) getTagValue(tagType string, field reflect.StructField) (string, bool) {
	tag := field.Tag.Get(tagType)
	if len(tag) > 0 {
		// get key value by splitting by comma because of tag options:
		// example: `json:"username,omitempty"` would return `username`
		return strings.Split(tag, ",")[0], true
	}
	return "", false
}
