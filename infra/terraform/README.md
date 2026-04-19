# Terraform infrastructure for `nfse`

This module provisions a minimal AWS footprint for running the `rodriguesflavio/nfse` container:

- **EC2** (Amazon Linux 2023) with Docker installed via `user_data`, pulling and running the image on `:8080` with `restart=unless-stopped`.
- **Security group** allowing inbound `:8080` from the internet (configurable) so callers can hit `/prestador` directly without the 30s API Gateway timeout.
- **IAM instance profile** with `AmazonSSMManagedInstanceCore` so you can shell in via SSM Session Manager without opening SSH.
- **HTTP API Gateway v2** (`enable_api_gateway = true` by default) with a catch-all `ANY /{proxy+}` route pointing at the EC2. Provided for convenience / future lightweight endpoints; **do not call `/prestador` through it** unless your flow completes in under 30 seconds.

## Why EC2 is the primary entry point

API Gateway HTTP API has a hard 30s integration timeout. The nfse `/prestador` flow takes ~25s for login alone and longer when emitting notas, so calling it through API Gateway would return 504 while the EC2 keeps processing. The EC2's public `:8080` is the intended entry point for `/prestador`; API Gateway is only useful for short requests.

## Prerequisites

- Terraform `>= 1.5`
- AWS credentials in the environment (`aws configure`, `aws sso login`, or `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` env vars). The identity used needs permission to create EC2 instances, security groups, IAM roles, instance profiles, and HTTP API Gateway v2 resources.

## Usage

### From GitHub Actions (recommended)

A manual-dispatch workflow at [`.github/workflows/terraform.yml`](../../.github/workflows/terraform.yml) runs `terraform plan`, `apply`, or `destroy` against this module. Trigger it from **Actions → Terraform → Run workflow**, pick an action, and hit **Run workflow**.

Before the first run, add two repo secrets at https://github.com/fvrodrigues/nfs/settings/secrets/actions:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

The IAM principal behind those keys needs permissions for EC2, VPC (read default), IAM (role + instance profile), API Gateway v2, and SSM (`GetParameter` for the AMI lookup).

**State persistence caveat.** The workflow uploads `terraform.tfstate` as a workflow artifact after every `apply`/`destroy` and downloads the most recent successful artifact before each run. This gives you working state across sequential dispatches but is **single-operator, not concurrent**. Don't trigger two `apply`s back-to-back, and don't mix local `terraform apply` with workflow `apply` — the artifact won't see local changes and vice versa. If more than one person needs to run this, switch to an S3 + DynamoDB backend.

### From your laptop

```bash
cd infra/terraform
terraform init
cp terraform.tfvars.example terraform.tfvars  # optional; override defaults here
terraform plan
terraform apply
```

After apply, the outputs include:

- `app_url` — `http://<ec2-public-dns>:8080`, the primary entry point
- `api_gateway_url` — HTTP API Gateway invoke URL (null if disabled)
- `instance_id` — useful for `aws ssm start-session --target <id>`

Test `/prestador` against EC2 directly:

```bash
curl -X POST -H 'Content-Type: application/json' \
  --data @payload.json \
  "$(terraform output -raw app_url)/prestador"
```

## Variables

See [`variables.tf`](variables.tf) for the full list. Highlights:

| Variable | Default | Purpose |
| --- | --- | --- |
| `region` | `us-east-1` | AWS region. |
| `instance_type` | `t3.medium` | Realistic minimum for Chromium + Go. |
| `docker_image` | `rodriguesflavio/nfse:latest` | Image the host pulls and runs. |
| `app_port` | `8080` | Host + container port. |
| `app_ingress_cidrs` | `["0.0.0.0/0"]` | Who can hit `:8080` on EC2. Tighten to your callers' CIDR if possible. |
| `ssh_ingress_cidrs` | `[]` | SSH ingress. Empty disables SSH; use SSM Session Manager instead. |
| `key_pair_name` | `""` | Existing EC2 key pair name, only needed if you want SSH. |
| `enable_api_gateway` | `true` | Set to `false` to skip the HTTP API entirely. |

## Shell access without SSH

```bash
aws ssm start-session --target "$(terraform output -raw instance_id)"
```

Once inside, `docker ps` and `docker logs nfse` show the running container.

## Updating the image

The container is started with `restart=unless-stopped` but does NOT auto-pull new tags. To roll out a new `:latest`:

```bash
aws ssm start-session --target "$(terraform output -raw instance_id)"
sudo docker pull rodriguesflavio/nfse:latest
sudo docker rm -f nfse
sudo docker run -d --name nfse --restart unless-stopped -p 8080:8080 -e PORT=8080 rodriguesflavio/nfse:latest
```

Or just re-run `terraform apply -replace=aws_instance.nfse` to recreate the host with a fresh `user_data` bootstrap.

## Tearing down

```bash
terraform destroy
```

## Caveats

- The module uses the account's **default VPC** to stay simple. Swap `network.tf` for an explicit VPC or the `terraform-aws-modules/vpc` module if you need isolation.
- Remote state is **not configured** (local `terraform.tfstate`). For team use, add an S3 backend.
- `app_ingress_cidrs = ["0.0.0.0/0"]` means anyone on the internet who guesses the EC2 DNS can POST to `/prestador`. Tighten it or add an application-layer auth token before running in production.
- Docker Hub pulls over `docker pull` aren't authenticated — if Docker Hub rate-limits the instance, add a login step in `user_data.sh.tftpl`.
