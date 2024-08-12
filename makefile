run:
	go run app/services/sales-api/main.go
run-fmt:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

tidy:
	go mod tidy

curl-test:
	curl -iL http://localhost:9000/v1