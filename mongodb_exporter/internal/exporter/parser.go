package exporter

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/shoplineapp/go-app/plugins/logger"
	"github.com/sirupsen/logrus"
)

type Parser struct {
	logger *logger.Logger
	events *Events
}

func (p *Parser) StrToInt32(val string) int32 {
	i, _ := strconv.Atoi(val)
	return int32(i)
}

func (p *Parser) StrToFloat32(val string) float32 {
	f, _ := strconv.ParseFloat(val, 32)
	return float32(f)
}

func (p *Parser) RegSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:]
	return result
}

func (p *Parser) ParseLogs(cluster string, server string, logs string) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "err": err}).Error("Failed to parse logs")
			p.events.Publish(EVENT_LOGS_ENTRIES_PARSED, cluster, server, []LogInfoQuery{})
		}
	}()

	p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "count": len(logs)}).Debug("Parsing logs")
	if len(logs) == 0 {
		p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "count": len(logs)}).Debug("No log to be processed, skipping")
		p.events.Publish(EVENT_LOGS_ENTRIES_PARSED, cluster, server, []LogInfoQuery{})
		return
	}

	f, fErr := os.CreateTemp("tmp/logs", "mongo-logs-")
	if fErr != nil {
		p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "error": fErr}).Error("Unable to create temporary file for logs")
	}
	f.WriteString(logs)

	// close and remove the temporary file at the end of the program
	defer f.Close()
	defer os.Remove(f.Name())

	var out bytes.Buffer
	var stderrr bytes.Buffer

	// mtools disallow call from non-TTY context and throw an error with "this tool can't parse input from stdin"
	// https://github.com/rueckstiess/mtools/issues/404
	// Skipping cluster info and table headers
	cmd := exec.Command("script", "-q", "-c", fmt.Sprintf("mloginfo --no-progressbar --queries %s | tail -n +15", f.Name()))
	cmd.Stdout = &out
	cmd.Stderr = &stderrr
	err := cmd.Run()
	if err != nil {
		p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "error": err}).Error("Error parsing logs")
	}

	entries := []LogInfoQuery{}

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	for scanner.Scan() {
		line := scanner.Text()

		// Columns are turned from tab into multiple spaces from stdout return
		parts := p.RegSplit(line, "[ ]{3,}")
		if len(parts) < 8 {
			continue
		}
		entries = append(entries, LogInfoQuery{
			Cluster:   cluster,
			Server:    server,
			Namespace: parts[0],
			Operation: parts[1],
			Pattern:   parts[2],
			Count:     p.StrToInt32(parts[3]),
			MinMS:     p.StrToFloat32(parts[4]),
			MaxMS:     p.StrToFloat32(parts[5]),
			P95MS:     p.StrToFloat32(parts[6]),
			SumMS:     p.StrToFloat32(parts[7]),
			MeanMS:    p.StrToFloat32(parts[8]),
		})
	}

	p.events.Publish(EVENT_LOGS_ENTRIES_PARSED, cluster, server, entries)
}

func NewParser(logger *logger.Logger, events *Events, store *Store) *Parser {
	p := &Parser{
		logger: logger,
		events: events,
	}
	p.events.SubscribeAsync(EVENT_LOGS_DOWNLOADED, p.ParseLogs, false)
	return p
}
