package shared

import "fmt"

func ErrLength(min, max int32) string {
	return fmt.Sprintf("Debe tener entre %d a %d caracteres", min, max)
}
