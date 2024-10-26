package helpers

import "time"

func GetDayMonthYearFrom(date string) string {
	convertedDate, err := time.Parse("2006-01-02 15:04", date)
	if err != nil {
		convertedDate = time.Now()
	}
	return convertedDate.Format("02.01.2006")
}
