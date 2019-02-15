run:
	go run main/main.go

test:
	go test ./...

image:
	docker build -t al-master .

docker-push:
	echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
	docker tag al-master codealife/al-master:latest
	docker push codealife/al-master:latest