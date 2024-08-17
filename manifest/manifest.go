package manifest

import (
	"fmt"
	"maps"
	"sort"
	"strings"
)

type Manifest map[string]Availability

type Availability struct {
	Price    uint
	Tag      string
	Url      string
	Currency string
}

func (m *Manifest) GenerateMessage() string {
	var builder strings.Builder

	keys := []string{}
	for key := range *m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, name := range keys {
		availability := (*m)[name]
		if availability.Tag == "" {
			fmt.Fprintf(&builder, "- ❌ %v: not found\n%v\n\n", name, availability.Url)
		} else {
			parts := strings.Split(availability.Tag, "/")
			fmt.Fprintf(
				&builder, "- %s %v: %v @ %.2f %s\n%v\n\n",
				generate_icon(availability.Tag), name, parts[len(parts)-1],
				float64(availability.Price)*0.01, availability.Currency, availability.Url,
			)
		}
	}

	return strings.TrimSpace(builder.String())

}

func AreEqual(m1, m2 Manifest) bool {
	if m1 == nil && m2 == nil {
		return true
	}
	if m1 == nil || m2 == nil {
		return false
	}
	return maps.Equal(m1, m2)
}

func generate_icon(tag string) string {
	switch tag {
	case "https://schema.org/OutOfStock":
		return "❌"
	case "https://schema.org/InStock":
		return "✅"
	default:
		return "⚠️"
	}
}
