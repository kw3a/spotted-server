package utils

import "fmt"

func MapToError(problems map[string]string) error {
  if len(problems) == 0 {
    return nil
  }
  return fmt.Errorf("problems: %v", problems)
}
