package reflect

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func (ref reflection) flat() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s %s %s\n", ref.Method, ref.URI, ref.Proto)
	fmt.Fprintf(&b, "Host: %s\n", ref.Host)

	for _, key := range slices.Sorted(maps.Keys(ref.Headers)) {
		for _, value := range ref.Headers[key] {
			fmt.Fprintf(&b, "%s: %s\n", key, value)
		}
	}

	if ref.BodyBytes > 0 {
		b.WriteString("\n")
		b.WriteString(ref.Body)
		b.WriteString("\n")
	}

	return b.String()
}
