package shared

import (
	"fmt"
	"strconv"
)

func ErrLength(min, max int32) string {
	return fmt.Sprintf("Debe tener entre %d a %d caracteres", min, max)
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
