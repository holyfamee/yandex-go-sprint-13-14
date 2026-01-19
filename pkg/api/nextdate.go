package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %w", err)
	}

	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", errors.New("некорректный формат правила")
	}

	rule := parts[0]

	switch rule {
	case "y":
		return handleYearlyRepeat(now, date)

	case "d":
		if len(parts) < 2 {
			return "", errors.New("не указан интервал в днях")
		}
		return handleDailyRepeat(now, date, parts[1])

	case "w":
		if len(parts) < 2 {
			return "", errors.New("не указаны дни недели")
		}
		return handleWeeklyRepeat(now, date, parts[1])

	case "m":
		if len(parts) < 2 {
			return "", errors.New("не указаны дни месяца")
		}
		months := ""
		if len(parts) >= 3 {
			months = parts[2]
		}
		return handleMonthlyRepeat(now, date, parts[1], months)

	default:
		return "", fmt.Errorf("неподдерживаемый формат правила: %s", rule)
	}
}

// afterNow проверяет, что дата больше now (не учитывая время)
func afterNow(date, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()

	if y1 > y2 {
		return true
	}
	if y1 == y2 && m1 > m2 {
		return true
	}
	if y1 == y2 && m1 == m2 && d1 > d2 {
		return true
	}
	return false
}

// handleYearlyRepeat обрабатывает правило ежегодного повторения
func handleYearlyRepeat(now, date time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(dateFormat), nil
}

// handleDailyRepeat обрабатывает правило повторения через N дней
func handleDailyRepeat(now, date time.Time, intervalStr string) (string, error) {
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return "", fmt.Errorf("некорректный интервал в днях: %w", err)
	}

	if interval <= 0 || interval > 400 {
		return "", errors.New("интервал должен быть от 1 до 400 дней")
	}

	for {
		date = date.AddDate(0, 0, interval)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(dateFormat), nil
}

// handleWeeklyRepeat обрабатывает правило повторения в указанные дни недели
func handleWeeklyRepeat(now, date time.Time, daysStr string) (string, error) {
	dayParts := strings.Split(daysStr, ",")
	var allowedDays [8]bool

	for _, dayStr := range dayParts {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			return "", fmt.Errorf("некорректный день недели: %w", err)
		}
		if day < 1 || day > 7 {
			return "", fmt.Errorf("день недели должен быть от 1 до 7, получено: %d", day)
		}
		allowedDays[day] = true
	}

	date = date.AddDate(0, 0, 1)

	for {
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7
		}

		if allowedDays[weekday] && afterNow(date, now) {
			break
		}
		date = date.AddDate(0, 0, 1)
	}

	return date.Format(dateFormat), nil
}

// handleMonthlyRepeat обрабатывает правило повторения в указанные дни месяца
func handleMonthlyRepeat(now, date time.Time, daysStr, monthsStr string) (string, error) {
	dayParts := strings.Split(daysStr, ",")
	var allowedDays [32]bool
	var negativeDays []int
	hasNegativeDays := false

	for _, dayStr := range dayParts {
		dayStr = strings.TrimSpace(dayStr)
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return "", fmt.Errorf("некорректный день месяца: %w", err)
		}

		if day < 0 {
			if day < -2 {
				return "", fmt.Errorf("отрицательный день должен быть -1 или -2, получено: %d", day)
			}
			negativeDays = append(negativeDays, day)
			hasNegativeDays = true
		} else {
			if day < 1 || day > 31 {
				return "", fmt.Errorf("день месяца должен быть от 1 до 31, получено: %d", day)
			}
			allowedDays[day] = true
		}
	}

	var allowedMonths [13]bool

	if monthsStr != "" {
		monthParts := strings.Split(monthsStr, ",")
		for _, monthStr := range monthParts {
			month, err := strconv.Atoi(strings.TrimSpace(monthStr))
			if err != nil {
				return "", fmt.Errorf("некорректный месяц: %w", err)
			}
			if month < 1 || month > 12 {
				return "", fmt.Errorf("месяц должен быть от 1 до 12, получено: %d", month)
			}
			allowedMonths[month] = true
		}
	} else {
		for i := 1; i <= 12; i++ {
			allowedMonths[i] = true
		}
	}

	date = date.AddDate(0, 0, 1)

	for {
		day := date.Day()
		month := int(date.Month())

		if !allowedMonths[month] {
			date = date.AddDate(0, 1, -day+1)
			continue
		}

		dayMatches := false

		if allowedDays[day] {
			dayMatches = true
		}

		if hasNegativeDays {
			daysInMonth := daysInMonth(date.Year(), month)
			for _, negDay := range negativeDays {
				targetDay := daysInMonth + negDay + 1
				if day == targetDay {
					dayMatches = true
					break
				}
			}
		}

		if dayMatches && afterNow(date, now) {
			break
		}

		date = date.AddDate(0, 0, 1)

		if date.Year() > now.Year()+2 {
			return "", errors.New("не удалось найти следующую дату")
		}
	}

	return date.Format(dateFormat), nil
}

func daysInMonth(year, month int) int {
	firstDayNextMonth := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	lastDayCurrentMonth := firstDayNextMonth.AddDate(0, 0, -1)
	return lastDayCurrentMonth.Day()
}

// nextDateHandler обрабатывает GET-запросы к /api/nextdate
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error
	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, "некорректный формат параметра now", http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}
