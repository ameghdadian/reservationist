package page

import (
	"fmt"
	"strconv"
)

const (
	maxRowsPerPage = 100
)

type Page struct {
	number int
	rows   int
}

func Parse(page string, rowsPerPage string) (Page, error) {
	number := 1
	if page != "" {
		var err error
		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, fmt.Errorf("page conversion: %w", err)
		}
	}

	rows := 10
	if rowsPerPage != "" {
		var err error
		rows, err = strconv.Atoi(rowsPerPage)
		if err != nil {
			return Page{}, fmt.Errorf("rows conversion: %w", err)
		}
	}

	if number <= 0 {
		return Page{}, fmt.Errorf("page value too small, must be larger than 0")
	}

	if rows <= 0 {
		return Page{}, fmt.Errorf("rows value too small, must be grater than 0")
	}

	if rows > maxRowsPerPage {
		return Page{}, fmt.Errorf("rows value too large, must be less than %d", maxRowsPerPage)
	}

	p := Page{
		number: number,
		rows:   rows,
	}

	return p, nil
}

// MustParse creates a paging value for testing.
func MustParse(page string, rowsPerPage string) Page {
	pg, err := Parse(page, rowsPerPage)
	if err != nil {
		panic(err)
	}

	return pg
}

func (p Page) String() string {
	return fmt.Sprintf("page: %d rows: %d", p.number, p.rows)
}

func (p Page) Number() int {
	return p.number
}

func (p Page) RowsPerPage() int {
	return p.rows
}
