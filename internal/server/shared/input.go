package shared

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var spanishMonths = map[string]string{
	"January":   "Enero",
	"February":  "Febrero",
	"March":     "Marzo",
	"April":     "Abril",
	"May":       "Mayo",
	"June":      "Junio",
	"July":      "Julio",
	"August":    "Agosto",
	"September": "Septiembre",
	"October":   "Octubre",
	"November":  "Noviembre",
	"December":  "Diciembre",
}

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

func DateSpanishFormat(t sql.NullTime) string {
	if !t.Valid {
		return "presente"
	}
	englishFormatted := t.Time.Format("January 2006")
	for en, es := range spanishMonths {
		englishFormatted = strings.ReplaceAll(englishFormatted, en, es)
	}
	return englishFormatted
}

func TimeInterval(start time.Time, end sql.NullTime) string {
	var duration time.Duration
	if end.Valid {
		duration = end.Time.Sub(start)
	} else {
		duration = time.Since(start)
	}

	if duration < 0 {
		return "menos de un mes"
	}

	years := int(duration.Hours() / (24 * 365))
	months := int((duration.Hours() / (24 * 30)) - float64(years*12))

	if years > 0 && months > 0 {
		return fmt.Sprintf("%d año%s %d mes%s", years, Pluralize(years, false), months, Pluralize(months, true))
	} else if years > 0 {
		return fmt.Sprintf("%d año%s", years, Pluralize(years, false))
	} else if months > 0 {
		return fmt.Sprintf("%d mes%s", months, Pluralize(months, true))
	}
	return "menos de un mes"
}

func Pluralize(value int, month bool) string {
	if value == 1 {
		return ""
	}
	if month {
		return "es"
	}
	return "s"
}

