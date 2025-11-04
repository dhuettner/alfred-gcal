//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2019-04-03
//

package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// quickAdd check if there are configured accounts and pass data to create an event.
func quickAdd() error {
	log.Println("Creating event", opts.Quick, opts.CalendarID)

	if err := createEvent(opts.Quick, opts.CalendarID); err != nil {
		return err
	}

	if err := doUpdateEvents(); err != nil {
		return err
	}

	return nil
}

// createEvent looks for account by calendar ID and create new event in that account.
// If the input starts with "f:" or "focus:", it creates a Focus Time event instead.
func createEvent(quick string, calendarID string) error {
	// Check if this is a Focus Time event
	if strings.HasPrefix(quick, "f:") || strings.HasPrefix(quick, "focus:") {
		return createFocusTimeEvent(quick, calendarID)
	}

	// Regular QuickAdd event
	for _, acc := range accounts {
		for _, c := range acc.Calendars {
			if c.ID == calendarID {
				return acc.QuickAdd(calendarID, quick)
			}
		}
	}

	return nil
}

// createFocusTimeEvent parses the input and creates a Focus Time event.
// Syntax: "f: Title [duration]" or "focus: Title [duration]"
// Examples:
//   - "f: Deep Work" -> 1h from now
//   - "f: Deep Work 2h" -> 2h from now
//   - "f: Deep Work 30m" -> 30 minutes from now
func createFocusTimeEvent(input string, calendarID string) error {
	// Remove prefix
	input = strings.TrimPrefix(input, "focus:")
	input = strings.TrimPrefix(input, "f:")
	input = strings.TrimSpace(input)

	// Parse duration at the end (e.g., "2h", "30m", "1.5h")
	duration := 1 * time.Hour // default duration
	title := input

	// Regex to match duration at the end: "2h", "30m", "1.5h", etc.
	durationRegex := regexp.MustCompile(`\s+(\d+(?:\.\d+)?)(h|m)$`)
	if matches := durationRegex.FindStringSubmatch(input); len(matches) > 0 {
		// Extract duration value and unit
		valueStr := matches[1]
		unit := matches[2]

		value, err := strconv.ParseFloat(valueStr, 64)
		if err == nil {
			switch unit {
			case "h":
				duration = time.Duration(value * float64(time.Hour))
			case "m":
				duration = time.Duration(value * float64(time.Minute))
			}
		}

		// Remove duration from title
		title = strings.TrimSpace(durationRegex.ReplaceAllString(input, ""))
	}

	// Find the account that owns this calendar
	// Focus Time events must be created on the primary calendar
	for _, acc := range accounts {
		// Try to find the calendar to get the account
		for _, c := range acc.Calendars {
			if c.ID == calendarID {
				// Create Focus Time event starting now
				start := time.Now().Local()
				log.Printf("[focustime] creating event: title=%q, start=%v, duration=%v", title, start, duration)
				return acc.CreateFocusTime(title, start, duration)
			}
		}
	}

	return nil
}
