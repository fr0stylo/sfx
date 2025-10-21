package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/provider"
)

type Options struct {
	Path string `yaml:"path"`
}

func main() {
	provider.Run(provider.HandlerFunc(handle))
}

func handle(req provider.Request) (provider.Response, error) {
	var opts Options
	if err := yaml.Unmarshal(req.Options, &opts); err != nil {
		return provider.Response{}, err
	}

	refUri, err := url.Parse(req.Ref)
	if err != nil {
		return provider.Response{}, fmt.Errorf("parse ref: %w", err)
	}

	f, err := os.Open(opts.Path)
	if err != nil {
		return provider.Response{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	switch refUri.Scheme {
	case "env":
		buf, err := parseEnvFile(f, []byte(refUri.Hostname()))
		return provider.Response{Value: buf}, err
	default:
		return provider.Response{}, fmt.Errorf("unsupported scheme %q", refUri.Scheme)
	}
}

func parseEnvFile(r io.Reader, ref []byte) ([]byte, error) {
	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		line := scan.Bytes()
		if bytes.HasPrefix(line, ref) {
			line = bytes.TrimPrefix(line, ref)
			line = bytes.Trim(line, " \t\r\n\"'=")
			return line, nil
		}
	}

	return nil, nil
}
