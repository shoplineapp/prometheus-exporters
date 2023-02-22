package legacy

import (
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
)

// https://github.com/rueckstiess/mtools/blob/c3f4721018854fff09712c5a17096d4cfbb4e3c7/mtools/mloginfo/sections/query_section.py#L87
var (
	ANALYZABLE_OPERATIONS = []string{"query", "getmore", "command", "write", "update", "remove"}
	ANALYZABLE_COMMANDS   = []string{"query", "getmore", "update", "remove", "count", "findandmodify", "geonear", "find", "aggregate", "command"}
)

func (l *LogConverter) ParseRawObjectString(sb *strings.Builder, v *fastjson.Object, field string) {
	bytes := []byte{}
	sb.WriteString(field)
	sb.WriteString(":")
	if v.Get(field) == nil {
		sb.WriteString("{}")
		sb.WriteString(" ")
		return
	}

	val := string(v.Get(field).GetObject().MarshalTo(bytes))
	if val == "" {
		val = "{}"
	}
	sb.WriteString(val)
	sb.WriteString(" ")
}

func (l *LogConverter) ParseStringField(sb *strings.Builder, v *fastjson.Object, field string, defaultValue string) {
	sb.WriteString(field)
	sb.WriteString(":")
	val := string(v.Get(field).GetStringBytes())
	if val == "" && defaultValue != val {
		val = defaultValue
	}
	sb.WriteString(val)
	sb.WriteString(" ")
}

func (l *LogConverter) ParseNumericField(sb *strings.Builder, v *fastjson.Object, field string) {
	sb.WriteString(field)
	sb.WriteString(":")
	f := v.Get(field).GetInt()
	sb.WriteString(strconv.Itoa(f))
	sb.WriteString(" ")
}

func (l *LogConverter) ParseCommand(sb *strings.Builder, v *fastjson.Object) {
	command := v.Get("command").GetObject()
	sb.WriteString(" command: ")
	for _, cmd := range ANALYZABLE_COMMANDS {
		if string(command.Get(cmd).GetStringBytes()) != "" {
			sb.WriteString(cmd)
			break
		}
	}
	sb.WriteString(" ")
	cb := []byte{}
	sb.WriteString(string(command.MarshalTo(cb)))
	sb.WriteString(" ")
}
