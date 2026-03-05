.PHONY: build run query test

build:
	go build -o adguard-home-exporter .

run:
	ADGUARD_USERNAME=$$(op read "op://Private/AdGuard Home/username") \
	ADGUARD_PASSWORD=$$(op read "op://Private/AdGuard Home/password") \
	go run .

query:
	curl -s http://localhost:9617/metrics

test:
	go test ./... -v
