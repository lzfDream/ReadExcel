package types

var BuildInType = map[string]struct{}{
	"bool":   {},
	"int":    {},
	"double": {},
	"string": {},
}

var CustomType = map[string]ClassDefine{}
var CustomTypeDetail = map[string][]ClassDefine{}

func init() {
	newBuildInType := make(map[string]struct{}, 2*len(BuildInType))
	for strType := range BuildInType {
		newBuildInType[strType] = struct{}{}
		newBuildInType[strType+"?"] = struct{}{}
	}
	BuildInType = newBuildInType
}

type Field struct {
	Name       string
	Type       string
	ForeignKey string
}

type ClassDefine struct {
	Name      string
	Separator string
	Fields    []Field
	// TODO: 类型扩展功能
}

func IsValid(strType string) bool {
	_, ok := BuildInType[strType]
	if !ok {
		_, ok = CustomType[strType]
	}
	return ok
}

func LoadFile(path string) error {
	test_type := ClassDefine{
		"ItemCount",
		",",
		[]Field{
			{"ID", "int", "item.id"},
			{"count", "int", ""},
		},
	}
	CustomType[test_type.Name] = test_type
	CustomTypeDetail["test"] = []ClassDefine{test_type}

	return nil
}

func ParseFile(rows [][]string) error {
	return nil
}
