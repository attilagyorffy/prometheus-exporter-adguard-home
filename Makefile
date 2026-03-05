.PHONY: build run query test

build:
	go build -o adguard-home-exporter .

run:
	ADGUARD_USERNAME=$$(op read "op://Private/kpk2enexqzl5ya2rrxkyfebehu/username") \
	ADGUARD_PASSWORD=$$(op read "op://Private/kpk2enexqzl5ya2rrxkyfebehu/password") \
	go run .

query:
	curl -s http://localhost:9617/metrics

test:
	go test ./... -v
