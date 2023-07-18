package ui

import (
	"fmt"
	"strings"
)

type Column struct {
	Title string
	Width int
}

type Row []string

func center(s string, w int) string {
	return fmt.Sprintf("%*s", -w, fmt.Sprintf("%*s", (w+len(s))/2, s))
}

func Table(title string, cols []Column, rows []Row) {
	totalWidth := 0
	header := ""
	body := ""

	for idx, col := range cols {
		if col.Width == -1 {
			// Use the widest row in the column
			for _, row := range rows {
				if len(row[idx]) > cols[idx].Width {
					cols[idx].Width = len(row[idx])
				}
			}
			cols[idx].Width += 2 // add some padding
		}
		// Set minimum column width to the column title
		if len(col.Title) > cols[idx].Width {
			cols[idx].Width = len(col.Title) + 2
		}

		totalWidth += cols[idx].Width
		header += fmt.Sprintf(" %-*s", -cols[idx].Width, cols[idx].Title)
	}
	header += "\n"
	title = center(title, totalWidth)
	separator := strings.Repeat("-", totalWidth+len(cols)+1) + "\n"

	for _, row := range rows {
		for idx, col := range cols {
			// Truncate the row to the column width
			if len(row[idx]) > col.Width {
				row[idx] = row[idx][:col.Width]
			}
			body += fmt.Sprintf(" %-*s", -col.Width, row[idx])
		}
		body += "\n"
	}

	fmt.Printf("%s\n%s%s%s%s", title, separator, header, separator, body)
}

// MaxFieldLength can be used to compute the maximum length of any given column based on the length of the greatest row
func MaxFieldLength[T any](data []T, getField func(T) string) int {
	maxLength := 0

	for _, item := range data {
		fieldValue := getField(item)
		if len(fieldValue) > maxLength {
			maxLength = len(fieldValue)
		}
	}

	return maxLength
}
