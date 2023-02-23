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
	UNQUOTABLE_KEYS       = []string{"filter", "aggregate", "filter", "pipeline", "distinct", "query"}
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
	fullCmd := v.Get("command").GetObject()
	cmdKey := ""

	sb.WriteString(" command: ")
	for _, cmd := range ANALYZABLE_COMMANDS {
		if string(fullCmd.Get(cmd).GetStringBytes()) != "" {
			cmdKey = cmd
			sb.WriteString(cmd)
			break
		}
	}
	sb.WriteString(" ")
	cb := []byte{}

	commandStr := string(fullCmd.MarshalTo(cb))

	// mtools does not work with keys with quotes and it's greping the pattern like "filter: "
	// We will remove the quotes and adding back trailing space after :
	// https://github.com/rueckstiess/mtools/blob/develop/mtools/util/logevent.py#L515
	if cmdKey != "" {
		commandStr = strings.ReplaceAll(commandStr, "\""+cmdKey+"\":", " "+cmdKey+": ")
		for _, key := range UNQUOTABLE_KEYS {
			commandStr = strings.ReplaceAll(commandStr, "\""+key+"\":", " "+key+": ")
		}
	}
	sb.WriteString(commandStr)
	sb.WriteString(" ")
}
