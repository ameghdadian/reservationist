package web

import "net/http"

func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}
