package commands

import (
	"encoding/json"
	"fmt"
)

func prettyPrint(model any) (string, error) {
	b, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to MarshalIndent the model: %w", err)
	}
	return string(b), nil
}
