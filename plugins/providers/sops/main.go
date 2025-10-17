package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/getsops/sops/v3/decrypt"
	"gopkg.in/yaml.v3"

	"sfx/provider"
)

type options struct {
	Path    string `yaml:"path"`
	Format  string `yaml:"format"`
	KeyPath string `yaml:"key_path"`
}

func main() {
	provider.Run(provider.HandlerFunc(handle))
}

func handle(req provider.Request) (provider.Response, error) {
	var opts options
	if len(req.Options) > 0 {
		if err := yaml.Unmarshal(req.Options, &opts); err != nil {
			return provider.Response{}, fmt.Errorf("parse options: %w", err)
		}
	}

	path, key := splitRef(req.Ref)
	if path == "" {
		path = opts.Path
	}
	if path == "" {
		return provider.Response{}, errors.New("ref must include a file path or options.path must be set")
	}

	if key == "" {
		key = opts.KeyPath
	}

	cleartext, err := decrypt.File(path, opts.Format)
	if err != nil {
		return provider.Response{}, fmt.Errorf("decrypt %q: %w", path, err)
	}

	if key == "" {
		return provider.Response{Value: cleartext}, nil
	}

	var root any
	if err := yaml.Unmarshal(cleartext, &root); err != nil {
		return provider.Response{}, fmt.Errorf("decode decrypted payload: %w", err)
	}

	value, err := navigate(root, parsePath(key))
	if err != nil {
		return provider.Response{}, err
	}

	buf, err := encodeValue(value)
	if err != nil {
		return provider.Response{}, err
	}

	return provider.Response{Value: buf}, nil
}

func splitRef(ref string) (string, string) {
	parts := strings.SplitN(ref, "#", 2)
	file := strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		return file, strings.TrimSpace(parts[1])
	}
	return file, ""
}

func parsePath(path string) []string {
	var (
		parts   []string
		current strings.Builder
		escape  bool
	)

	for _, r := range path {
		switch {
		case escape:
			current.WriteRune(r)
			escape = false
		case r == '\\':
			escape = true
		case r == '.' || r == '/':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func navigate(node any, path []string) (any, error) {
	if len(path) == 0 {
		return node, nil
	}

	head, tail := path[0], path[1:]

	switch val := node.(type) {
	case map[string]any:
		child, ok := val[head]
		if !ok {
			return nil, fmt.Errorf("key %q not found", head)
		}
		return navigate(child, tail)
	case map[any]any:
		if child, ok := val[head]; ok {
			return navigate(child, tail)
		}
		for k, v := range val {
			if ks, ok := k.(string); ok && ks == head {
				return navigate(v, tail)
			}
		}
		return nil, fmt.Errorf("key %q not found", head)
	case []any:
		idx, err := strconv.Atoi(head)
		if err != nil {
			return nil, fmt.Errorf("expected numeric index, got %q", head)
		}
		if idx < 0 || idx >= len(val) {
			return nil, fmt.Errorf("index %d out of range", idx)
		}
		return navigate(val[idx], tail)
	default:
		return nil, fmt.Errorf("cannot descend into %T for segment %q", node, head)
	}
}

func encodeValue(v any) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	case fmt.Stringer:
		return []byte(t.String()), nil
	case json.RawMessage:
		return []byte(t), nil
	case []any, map[string]any, map[any]any:
		normalized := normalize(t)
		buf, err := json.Marshal(normalized)
		if err != nil {
			return nil, fmt.Errorf("marshal value: %w", err)
		}
		return buf, nil
	default:
		buf, err := json.Marshal(t)
		if err != nil {
			return []byte(fmt.Sprint(t)), nil
		}
		return buf, nil
	}
}

func normalize(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, sub := range val {
			out[k] = normalize(sub)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(val))
		for k, sub := range val {
			key := fmt.Sprint(k)
			out[key] = normalize(sub)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, sub := range val {
			out[i] = normalize(sub)
		}
		return out
	default:
		return val
	}
}
