.PHONY: build run query test

build:
	go build -o adguard-home-exporter .

run:
	ADGUARD_USERNAME=$$(op read "op://Agent Services/qyxfkofbnj2mee5rbrnwnotj5q/username") \
	ADGUARD_PASSWORD=$$(op read "op://Agent Services/qyxfkofbnj2mee5rbrnwnotj5q/password") \
	go run .

query:
	curl -s http://localhost:9617/metrics

test:
	go test ./... -v
