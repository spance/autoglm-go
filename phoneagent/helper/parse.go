package helper

import (
	"errors"
	"fmt"
	"strings"
)

type Action map[string]any

type ActionResult struct {
	Success              bool
	ShouldFinish         bool
	Message              string
	RequiresConfirmation bool
}

func ParseAction(response string) (Action, error) {
	fmt.Printf("Parsing action: %s\n", response)

	response = strings.TrimSpace(response)

	// case 1: do(action="Type" ...) / do(action="Type_Name" ...)
	if strings.HasPrefix(response, `do(action="Type"`) ||
		strings.HasPrefix(response, `do(action="Type_Name"`) {

		text, err := extractQuotedArg(response, "text")
		if err != nil {
			return nil, err
		}

		return Action{
			"_metadata": "do",
			"action":    "Type",
			"text":      text,
		}, nil
	}

	// case 2: generic do(...)
	if strings.HasPrefix(response, "do(") {
		action, err := parseDoCall(response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse do() action: %w", err)
		}
		return action, nil
	}

	// case 3: finish(message="...")
	if strings.HasPrefix(response, "finish") {
		msg, err := extractQuotedArg(response, "message")
		if err != nil {
			return nil, err
		}

		return Action{
			"_metadata": "finish",
			"message":   msg,
		}, nil
	}

	return nil, fmt.Errorf("failed to parse action: %s", response)
}

func parseDoCall(expr string) (Action, error) {
	// 去掉 do( 和 )
	if !strings.HasPrefix(expr, "do(") || !strings.HasSuffix(expr, ")") {
		return nil, errors.New("invalid do() syntax")
	}

	body := strings.TrimSuffix(strings.TrimPrefix(expr, "do("), ")")

	action := Action{
		"_metadata": "do",
	}

	if strings.TrimSpace(body) == "" {
		return action, nil
	}

	// parts := splitArgs(body)
	parts := strings.Split(body, ", ")

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid argument: %s", part)
		}

		key := strings.TrimSpace(kv[0])
		valStr := strings.TrimSpace(kv[1])

		val, err := parseLiteral(valStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %w", key, err)
		}

		action[key] = val
	}
	return action, nil
}

func extractQuotedArg(s, key string) (string, error) {
	idx := strings.Index(s, key+"=")
	if idx == -1 {
		return "", fmt.Errorf("missing %s", key)
	}

	rest := s[idx+len(key)+1:]
	if len(rest) < 2 || rest[0] != '"' {
		return "", fmt.Errorf("invalid %s format", key)
	}

	rest = rest[1:]
	end := strings.LastIndex(rest, `"`)

	if end == -1 {
		return "", fmt.Errorf("unterminated string for %s", key)
	}

	return rest[:end], nil
}

func splitArgs(s string) []string {
	var (
		args     []string
		current  strings.Builder
		inQuotes bool
	)

	for _, r := range s {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case ',':
			if inQuotes {
				current.WriteRune(r)
			} else {
				args = append(args, strings.TrimSpace(current.String()))
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}
	return args
}

func parseLiteral(s string) (any, error) {
	fmt.Printf("begin to parse literal: %s\n", s)
	// string
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return s[1 : len(s)-1], nil
	}

	// bool
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	// int[]
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		content := strings.TrimSpace(s[1 : len(s)-1])
		if content == "" {
			return []int{}, nil
		}

		parts := strings.Split(content, ",")
		result := make([]int, 0, len(parts))

		for _, p := range parts {
			p = strings.TrimSpace(p)
			var v int
			if _, err := fmt.Sscanf(p, "%d", &v); err != nil {
				return nil, fmt.Errorf("invalid int in array: %s", p)
			}
			result = append(result, v)
		}
		return result, nil
	}
	// int
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
		return i, nil
	}

	// float
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
		return f, nil
	}

	return nil, fmt.Errorf("unsupported literal: %s", s)
}

func Do(kwargs map[string]any) Action {
	kwargs["_metadata"] = "do"
	return kwargs
}

func Finish(kwargs map[string]any) Action {
	kwargs["_metadata"] = "finish"
	return kwargs
}
