all: dev
dev:
	go run *.go
docker-build:
	sudo docker build -t api *.go
docker-run:
	sudo docker run -d -p 8080:8080 api
generate-env:
	kubectl create secret generic api-env --from-file=.env -n api
clean:
	docker rm $(docker stop $(docker ps -a -q --filter ancestor=api --format="{{.ID}}"))
