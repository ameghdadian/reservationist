run:
	go run app/services/reservations-api/main.go

run-fmt:
	go run app/services/reservations-api/main.go | go run app/tooling/logfmt/main.go

tidy:
	go mod tidy

curl-test:
	curl -iL http://localhost:3000/v1

curl-auth:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1 

generate-token:
	go run app/tooling/reservations-admin/main.go --command gentoken
generate-private-key:
	go run app/tooling/reservations-admin/main.go --command genkey
generate-migrate-seed:
	go run app/tooling/reservations-admin/main.go --command migrateseed
