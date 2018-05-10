package dates

import (
	"fmt"
	"strconv"
	"time"

	"github.com/snabb/isoweek"
)

type IsoWeek struct {
	Year int
	Week int
}

func LastIsoWeek() IsoWeek {
	year, week := time.Now().ISOWeek()
	return IsoWeek{Year: year, Week: week}
}

func getIsoWeek(year int, week int) *IsoWeek {
	// This company was founded in 1993. No need to have snippets
	// from before that date.
	if year < 1993 {
		return nil
	}

	// Weeks in the future cannot be instantiated.
	currentYear, currentWeek := time.Now().ISOWeek()
	if year > currentYear || (year == currentYear && week > currentWeek) {
		return nil
	}

	return &IsoWeek{Year: year, Week: week}
}

func ParseIsoWeek(year string, week string) *IsoWeek {
	parsedYear, err := strconv.ParseInt(year, 10, 0)
	if err != nil {
		return nil
	}
	parsedWeek, err := strconv.ParseInt(week, 10, 0)
	if err != nil {
		return nil
	}
	if !isoweek.Validate(int(parsedYear), int(parsedWeek)) {
		return nil
	}
	return getIsoWeek(int(parsedYear), int(parsedWeek))
}

func (iw IsoWeek) String() string {
	return fmt.Sprintf("%4d-W%02d", iw.Year, iw.Week)
}

func (iw IsoWeek) Seek(weeks int) *IsoWeek {
	year, month, day := isoweek.StartDate(iw.Year, iw.Week)
	year, week := time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Add(time.Duration(weeks) * 7 * 24 * time.Hour).ISOWeek()
	return getIsoWeek(year, week)
}

func (iw IsoWeek) FirstDay() string {
	year, month, day := isoweek.StartDate(iw.Year, iw.Week)
	return fmt.Sprintf("%4d-%02d-%02d", year, month, day)
}

func (iw IsoWeek) LastDay() string {
	year, month, day := isoweek.StartDate(iw.Year, iw.Week)
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Add(6 * 24 * time.Hour)
	return fmt.Sprintf("%4d-%02d-%02d", t.Year(), t.Month(), t.Day())
}
