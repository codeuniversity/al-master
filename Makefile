run:
	go run main/main.go

test:
	go test ./...

image:
	docker build -t al-master .
