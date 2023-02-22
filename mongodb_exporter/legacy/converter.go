package legacy

import (
	"bufio"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
	"golang.org/x/exp/slices"
)

type LogConverter struct {
}

func (l *LogConverter) ParseFile(logFilePath string, destFilePath *string) {
	fmt.Println("====== ParseFile??", logFilePath)
	f, err := os.Open(logFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var dest *os.File
	if destFilePath != nil {
		dest, err = os.OpenFile(*destFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer dest.Close()
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		str := l.ConstructLine(scanner.Text())
		// dest.WriteString(sb.String())
		if str == "" {
			continue
		}
		if destFilePath != nil && dest != nil {
			dest.WriteString(str)
			dest.WriteString("\n")
		} else {
			fmt.Println(str)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func (l *LogConverter) ConstructLine(line string) string {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println("Failed to ConstructLine", err)
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			return
		}
	}()
	var sb strings.Builder
	var p fastjson.Parser
	v, _ := p.Parse(line)
	command := string(v.GetStringBytes("c"))
	if !slices.Contains(ANALYZABLE_OPERATIONS, strings.ToLower(command)) {
		return ""
	}

	attr := v.GetObject("attr")
	t := string(attr.Get("type").GetStringBytes())
	if !slices.Contains(ANALYZABLE_COMMANDS, t) {
		return ""
	}

	sb.WriteString(string(v.GetStringBytes("t", "$date")))
	sb.WriteString(" ")
	sb.WriteString(string(v.GetStringBytes("s")))
	sb.WriteString("  ")
	sb.WriteString(command)
	sb.WriteString("  [")
	sb.WriteString(string(v.GetStringBytes("ctx")))
	sb.WriteString("] ")
	sb.WriteString(t)
	sb.WriteString(" ")
	sb.WriteString(string(attr.Get("ns").GetStringBytes()))
	sb.WriteString(" ")
	l.ParseCommand(&sb, attr)
	l.ParseStringField(&sb, attr, "planSummary", "UNKNOWN {}")
	l.ParseNumericField(&sb, attr, "keysExamined")
	l.ParseNumericField(&sb, attr, "docsExamined")
	l.ParseNumericField(&sb, attr, "cursorExhausted")
	l.ParseNumericField(&sb, attr, "numYields")
	l.ParseNumericField(&sb, attr, "numYields")
	l.ParseNumericField(&sb, attr, "nreturned")
	l.ParseStringField(&sb, attr, "queryHash", "NOT_PROVIDED")
	l.ParseStringField(&sb, attr, "planCacheKey", "NOT_PROVIDED")
	l.ParseNumericField(&sb, attr, "reslen")
	l.ParseRawObjectString(&sb, attr, "locks")
	l.ParseRawObjectString(&sb, attr, "storage")
	l.ParseStringField(&sb, attr, "protocol", "NOT_PROVIDED")

	f := attr.Get("durationMillis").GetInt()
	sb.WriteString(strconv.Itoa(f))
	sb.WriteString("ms")
	return sb.String()
}
