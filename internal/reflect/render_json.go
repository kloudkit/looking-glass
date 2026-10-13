package reflect

import (
	"bytes"
	"encoding/json"
)

func (ref reflection) json() string {
	var b bytes.Buffer

	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(ref)

	return b.String()
}
