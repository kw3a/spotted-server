package server

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

func ValidateUUID(id string) error {
	return uuid.Validate(id)
}

func ValidateLanguageID(languageID string) (int32, error) {
	languageIDInt, err := strconv.ParseInt(languageID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("languageID is not a valid integer")
	}
	languageIDInt32 := int32(languageIDInt)
	if languageIDInt32 < 0 || languageIDInt32 > 100 {
		return 0, fmt.Errorf("languageID is not in the valid range")
	}
	return languageIDInt32, nil
}
