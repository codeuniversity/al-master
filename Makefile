run:
	go run main/main.go

test:
	go test ./...

image:
	docker build -t al-master .

pushImage:
    echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
    docker tag al-master monteymontey/al-master:latest
    docker push monteymontey/al-master:latest