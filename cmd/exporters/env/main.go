package main

import (
	"sort"
	"strconv"
	"strings"

	"sfx/exporter"
)

func main() {
	exporter.Run(exporter.HandlerFunc(handle))
}

func handle(req exporter.Request) (exporter.Response, error) {
	var keys []string
	for k := range req.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(formatValue(req.Values[k]))
	}
	b.WriteByte('\n')

	return exporter.Response{Payload: []byte(b.String())}, nil
}

func formatValue(value []byte) string {
	s := string(value)
	if s == "" {
		return `""`
	}

	if strings.ContainsAny(s, " \t\r\n\"'\\") {
		return strconv.Quote(s)
	}
	return s
}
