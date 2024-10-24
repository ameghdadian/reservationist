package page

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ameghdadian/service/foundation/validate"
)

const (
	maxRowsPerPage = 100
)

type Page struct {
	Number      int
	RowsPerPage int
}

func Parse(r *http.Request) (Page, error) {
	values := r.URL.Query()

	number := 1
	if page := values.Get("page"); page != "" {
		var err error
		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, validate.NewFieldsError("page", err)
		}
	}

	rowsPerPage := 10
	if rows := values.Get("rows"); rows != "" {
		var err error
		rowsPerPage, err = strconv.Atoi(rows)
		if err != nil {
			return Page{}, validate.NewFieldsError("rows", err)
		}
		if rowsPerPage > maxRowsPerPage {
			return Page{}, validate.NewFieldsError(
				"rows",
				fmt.Errorf("rows per page exceeded the limit: GOT: %d, MAX: %d", rowsPerPage, maxRowsPerPage),
			)
		}
	}

	return Page{
		Number:      number,
		RowsPerPage: rowsPerPage,
	}, nil
}
