# database setup

- create the cluster:
  - `helm install mongo -f dbconf.yaml bitnami/mongodb`
  - `--set mongodbRootPassword=secretpassword,mongodbUsername=my-user,mongodbPassword=my-password,mongodbDatabase=my-database` for specifying the environment through cli
  - see [this](https://github.com/bitnami/charts/tree/master/upstreamed/mongodb) for more info
  - see [this](https://github.com/bitnami/bitnami-docker-mongodb) for info on the docker image used
- connect to cluster from cli
  - `kubectl run --namespace default mongo-mongodb-client --rm --tty -i --restart='Never' --image bitnami/mongodb --command -- mongo website --host`
  - `mongo-mongodb --authenticationDatabase website -u user -p pass`
  - note: port forwarding just is not working. therefore mongodb compass doesn't work. need to go to the cluster manually as shown above.
- another option is to use the [official mongodb operator](https://docs.mongodb.com/kubernetes-operator/master/tutorial/install-k8s-operator/)
