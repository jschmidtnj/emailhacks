# search engine setup

- create the cluster:
  - `helm install elastic -f searchconf.yaml bitnami/elastic`
  - see some docs [here](https://github.com/bitnami/charts/tree/master/bitnami/elasticsearch)
- docker [image used](https://github.com/bitnami/bitnami-docker-elasticsearch)
- can potentially use [official elasticsearch k8s operator](https://www.elastic.co/guide/en/cloud-on-k8s/current/k8s-quickstart.html)
