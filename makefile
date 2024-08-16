run:
	go run app/services/sales-api/main.go

run-fmt:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

tidy:
	go mod tidy

curl-test:
	curl -iL http://localhost:9000/v1

curl-auth:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:9000/v1 

generate-token:
	go run app/tooling/sales-admin/main.go
