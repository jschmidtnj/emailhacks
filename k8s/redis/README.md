# redis setup

- used for pub sub and cache
- create the cluster:
  - `helm install redis -f redisconf.yaml bitnami/redis`
  - see some docs [here](https://github.com/bitnami/charts/tree/master/upstreamed/redis)
- docker [image used](https://github.com/bitnami/bitnami-docker-redis)
- can potentially use [official redis k8s operator](https://github.com/RedisLabs/redis-enterprise-k8s-docs)
