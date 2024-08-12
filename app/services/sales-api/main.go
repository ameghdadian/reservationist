package main

import (
	"os"

	"github.com/ameghdadian/service/app/services/sales-api/v1/cmd"
	"github.com/ameghdadian/service/app/services/sales-api/v1/cmd/all"
)

/*
	Next Steps:
		1. Logger fmt 						DONE
		2. Middlewares(start with Logger)
*/

/*
	Requirements:
	1. We need following signature for our handlers. They must be accepting context
	   and return an error. Inside the program, errors needs to be propagated up the chain.
	   This is required as we need to integrate error handling logic in a single
	   place which is inside the Error middleware.
	2. Neither http package HandlerFunc nor httptreemux provides the signature. We need
	   a little bit of customization to satisfy our requirements.

	func (ctx context.Context, w http.ResponseWriter, r *http.Request) error
*/

var build = "develop"
var routes = "all"

func main() {
	switch routes {
	case "all":
		if err := cmd.Main(build, all.Routes()); err != nil {
			os.Exit(1)
		}
	}
}
