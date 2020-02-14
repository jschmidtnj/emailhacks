# deploy to aws using terraform

- provide terraform config files here.

## using cloudformation

- https://github.com/nathanpeck/aws-cloudformation-fargate

## using terraform

- https://github.com/hashicorp/terraform/issues/62
- https://github.com/duduribeiro/terraform_ecs_fargate_example

### deploy to eks

- secrets - https://www.terraform.io/docs/providers/aws/r/secretsmanager_secret_version.html

### encrypt secret files

- `tar cvf secrets.tar .env serviceaccount.json`
- `travis encrypt-file secrets.tar`

app engine: `gcloud app deploy`

## add to production

### deploy to AWS Fargate

#### manually:

- cli [user guide](https://docs.aws.amazon.com/cli/latest/userguide/aws-cli.pdf)
- `aws secretsmanager create-secret --region us-east-1 --name SECRET --secret-string <value>`
- `sudo docker build -t mailpear:latest .`
- `sudo docker run --network="host" mailpear:latest`
- `sudo docker tag mailpear:latest <id>.dkr.ecr.us-east-1.amazonaws.com/mailpear-api:latest`
- `aws ecr get-login-password | sudo docker login --username AWS --password-stdin <id>.dkr.ecr.us-east-1.amazonaws.com/mailpear-api`
- `sudo docker push <id>.dkr.ecr.us-east-1.amazonaws.com/mailpear-api:latest`
- create [access key](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
- [see this](https://aws.amazon.com/blogs/compute/securing-credentials-using-aws-secrets-manager-with-aws-fargate/)

## edit CORS

The file `cors.json` is needed to allow file downloads from Firebase Storage. To configure CORS, download the [gsutil utility](https://cloud.google.com/storage/docs/gsutil_install) on your computer (again, linux preferred), and run `gcloud init` to sign in. Then run `export BOTO_CONFIG=/dev/null` on linux to prevent a bug in the program. Finally, run `gsutil cors set cors.json gs://<your-cloud-storage-bucket>` in the parent directory to add the CORS rules, obviously changing the command for your storage bucket. On windows, open google cloud sheel and run the commands above.
