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
// If the input starts with "a:" or "out:", it creates an Out of Office event instead.
func createEvent(quick string, calendarID string) error {
	// Check if this is an Out of Office event
	if strings.HasPrefix(quick, "a:") || strings.HasPrefix(quick, "out:") {
		return createOutOfOfficeEvent(quick, calendarID)
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

// createOutOfOfficeEvent parses the input and creates an Out of Office (Außer Haus) event.
// Syntax: "a: Title [duration]" or "out: Title [duration]"
// Examples:
//   - "a: Arzttermin" -> 1h from now
//   - "a: Außer Haus 2h" -> 2h from now
//   - "out: Lunch 30m" -> 30 minutes from now
func createOutOfOfficeEvent(input string, calendarID string) error {
	// Remove prefix
	input = strings.TrimPrefix(input, "out:")
	input = strings.TrimPrefix(input, "a:")
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
	// Out of Office events must be created on the primary calendar
	for _, acc := range accounts {
		// Try to find the calendar to get the account
		for _, c := range acc.Calendars {
			if c.ID == calendarID {
				// Create Out of Office event starting now
				start := time.Now().Local()
				log.Printf("[outofoffice] creating event: title=%q, start=%v, duration=%v", title, start, duration)
				return acc.CreateOutOfOffice(title, start, duration)
			}
		}
	}

	return nil
}
