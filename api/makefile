all: dev
dev:
	go run *.go
docker-build:
	sudo docker build -t go-docker .
docker-run:
	sudo docker run -d -p 8080:8080 go-docker
deploy:
	gcloud app deploy
clean:
	docker rm $(docker stop $(docker ps -a -q --filter ancestor=go-docker --format="{{.ID}}"))
