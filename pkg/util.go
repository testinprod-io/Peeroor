package pkg

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// BoolString is a custom type to handle both boolean and string representations.
type BoolString bool

// UnmarshalJSON implements custom unmarshaling for BoolString.
func (b *BoolString) UnmarshalJSON(data []byte) error {
	// Try as a boolean first.
	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		*b = BoolString(boolVal)
		return nil
	}

	// Next, try as a string.
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		// Trim whitespace and parse.
		strVal = strings.TrimSpace(strVal)
		parsed, err := strconv.ParseBool(strVal)
		if err != nil {
			return fmt.Errorf("invalid boolean string: %s", strVal)
		}
		*b = BoolString(parsed)
		return nil
	}

	return fmt.Errorf("BoolString: could not unmarshal %s", string(data))
}
