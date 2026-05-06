package annotations

import (
	"strings"
)

func Parse(text string) (name string, args map[string]string, ok bool) {
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimSpace(text)

	if !strings.HasPrefix(text, "#[") {
		return "", nil, false
	}

	text = strings.TrimPrefix(text, "#[")
	closingIdx := strings.LastIndexByte(text, ']')
	if closingIdx < 0 {
		return "", nil, false
	}
	text = text[:closingIdx]

	parenIdx := strings.IndexByte(text, '(')
	if parenIdx < 0 {
		return text, nil, true
	}

	name = text[:parenIdx]
	text = strings.TrimSuffix(text[parenIdx+1:], ")")
	args = parseNamedArgs(text)

	return name, args, true
}

func parseNamedArgs(s string) map[string]string {
	result := make(map[string]string)
	i := 0
	for i < len(s) {
		for i < len(s) && s[i] == ' ' {
			i++
		}
		if i >= len(s) {
			break
		}
		eq := strings.IndexByte(s[i:], '=')
		if eq < 0 {
			break
		}
		eq += i
		key := strings.TrimSpace(s[i:eq])
		i = eq + 1
		for i < len(s) && s[i] == ' ' {
			i++
		}
		if i >= len(s) {
			break
		}
		if s[i] == '{' {
			closing := strings.IndexByte(s[i:], '}')
			if closing < 0 {
				break
			}
			closing += i
			result[key] = s[i+1 : closing]
			i = closing + 1
		} else if s[i] == '"' {
			i++
			end := strings.IndexByte(s[i:], '"')
			if end < 0 {
				break
			}
			end += i
			result[key] = s[i:end]
			i = end + 1
		} else {
			end := i
			for end < len(s) && s[end] != ',' && s[end] != ' ' {
				end++
			}
			result[key] = strings.TrimSpace(s[i:end])
			i = end
		}
		if i < len(s) && s[i] == ',' {
			i++
		}
	}
	return result
}

func SplitList(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func TypeToServiceID(name string) string {
	var parts []string
	var current strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			parts = append(parts, strings.ToLower(current.String()))
			current.Reset()
		}
		current.WriteRune(r)
	}
	parts = append(parts, strings.ToLower(current.String()))
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, "_")
}

func ReceiverName(expr string) string {
	expr = strings.TrimPrefix(expr, "*")
	return expr
}

// RouteAnnotation is the serialized form of #[Route] for JSON output
type RouteAnnotation struct {
	Controller string   `json:"controller"`
	Action     string   `json:"action"`
	Path       string   `json:"path"`
	Methods    []string `json:"methods"`
	Name       string   `json:"name"`
	Public     bool     `json:"public"`
}

// RateLimitAnnotation is the serialized form of #[RateLimit] for JSON output
type RateLimitAnnotation struct {
	Controller string `json:"controller"`
	Action     string `json:"action"`
	Max        int    `json:"max"`
	Per        string `json:"per"`
}

// EntityAnnotation is the serialized form of #[Entity] for JSON output
type EntityAnnotation struct {
	StructName string `json:"struct_name"`
	Table      string `json:"table"`
	Repository string `json:"repository"`
}

// ValidateAnnotation is the serialized form of #[Validate] for JSON output
type ValidateAnnotation struct {
	Controller string `json:"controller"`
	Action     string `json:"action"`
	Rules      string `json:"rules"`
	Groups     string `json:"groups"`
}

// SubscribeAnnotation is the serialized form of #[Subscribe] for JSON output
type SubscribeAnnotation struct {
	Controller string `json:"controller"`
	Action     string `json:"action"`
	Event      string `json:"event"`
	Priority   int    `json:"priority"`
}

// CacheAnnotation is the serialized form of #[Cache] for JSON output
type CacheAnnotation struct {
	Controller string `json:"controller"`
	Action     string `json:"action"`
	TTL        int    `json:"ttl"`
	Key        string `json:"key"`
}

// SecurityAnnotation is the serialized form of #[Security] for JSON output
type SecurityAnnotation struct {
	Controller string   `json:"controller"`
	Action     string   `json:"action"`
	Role       string   `json:"role"`
	Roles      []string `json:"roles"`
	Strategy   string   `json:"strategy"`
	Key        string   `json:"key"`
}

// TemplateAnnotation is the serialized form of #[Template] for JSON output
type TemplateAnnotation struct {
	Controller string `json:"controller"`
	Action     string `json:"action"`
	Template   string `json:"template"`
}
