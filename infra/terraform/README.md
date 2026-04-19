# Terraform infrastructure for `nfse`

This module provisions a minimal AWS footprint for running the `rodriguesflavio/nfse` container:

- **EC2** (Amazon Linux 2023) with Docker installed via `user_data`, pulling and running the image on `:8080` with `restart=unless-stopped`.
- **Security group** allowing inbound `:8080` from the internet (configurable).

Callers hit `/prestador` directly on the EC2 host's public DNS. No load balancer or API Gateway is provisioned — the `/prestador` flow takes ~25s for login alone and longer when emitting notas, which exceeds API Gateway's 30s integration timeout.

## Prerequisites

- Terraform `>= 1.5`
- AWS credentials in the environment (`aws configure`, `aws sso login`, or `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` env vars). The identity used needs permission to create EC2 instances and security groups.

## Usage

### From GitHub Actions (recommended)

A manual-dispatch workflow at [`.github/workflows/terraform.yml`](../../.github/workflows/terraform.yml) runs `terraform plan`, `apply`, or `destroy` against this module. Trigger it from **Actions → Terraform → Run workflow**, pick an action, and hit **Run workflow**.

Before the first run, add two repo secrets at https://github.com/fvrodrigues/nfs/settings/secrets/actions:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

The IAM principal behind those keys needs permissions for EC2, VPC (read default), and SSM (`GetParameter` for the AMI lookup).

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

- `app_url` — `http://<ec2-public-dns>:8080`
- `instance_id` — EC2 instance ID
- `instance_public_dns` / `instance_public_ip` — raw host details

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
| `instance_type` | `t2.micro` | Sandbox/free-tier default. **Too small for the real workload** — Chromium will OOM on 1 GB RAM. Use `t3.medium` or larger for a functional deployment. |
| `docker_image` | `rodriguesflavio/nfse:latest` | Image the host pulls and runs. |
| `app_port` | `8080` | Host + container port. |
| `app_ingress_cidrs` | `["0.0.0.0/0"]` | Who can hit `:8080` on EC2. Tighten to your callers' CIDR if possible. |
| `ssh_ingress_cidrs` | `["0.0.0.0/0"]` | SSH ingress. Wide open by default so EC2 Instance Connect (browser SSH) works out of the box; restrict to your own CIDR in production. |
| `key_pair_name` | `""` | Existing EC2 key pair name. Optional — **EC2 Instance Connect does not need a key pair**; you only need one for `ssh -i` from your laptop. |

## Shell access

### EC2 Instance Connect (browser, no key pair)

1. AWS Console → EC2 → Instances → select the `nfse` instance.
2. **Connect** → **EC2 Instance Connect** tab → use `ec2-user`, click **Connect**.

A browser terminal opens directly on the host. Useful snippets:

```bash
sudo docker ps                          # is the container up?
sudo docker logs --tail 200 nfse        # recent logs
sudo docker logs -f nfse                # tail live
free -m                                 # memory pressure check (t2.micro has ~1 GB)
sudo journalctl -u docker --no-pager -n 50   # docker daemon health
```

If you see `Killed` / `OOMKilled` in `docker logs`, the instance doesn't have enough RAM to run Chromium — bump `instance_type` to `t3.medium` or larger.

### SSH from your laptop (optional)

Requires `key_pair_name` to be set to the name of an existing EC2 key pair and the matching `.pem` on your machine:

```bash
ssh -i /path/to/key.pem ec2-user@$(terraform output -raw instance_public_dns)
```

## Updating the image

The container is started with `restart=unless-stopped` but does NOT auto-pull new tags. To roll out a new `:latest`, re-run the Terraform workflow with `action=apply` (or run `terraform apply -replace=aws_instance.nfse` locally) — this recreates the EC2 host with a fresh `user_data` bootstrap that re-pulls the image.

## Tearing down

```bash
terraform destroy
```

## Caveats

- The module uses the account's **default VPC** to stay simple. Swap `network.tf` for an explicit VPC or the `terraform-aws-modules/vpc` module if you need isolation.
- Remote state is **not configured** (local `terraform.tfstate`). For team use, add an S3 backend.
- `app_ingress_cidrs = ["0.0.0.0/0"]` means anyone on the internet who guesses the EC2 DNS can POST to `/prestador`. Tighten it or add an application-layer auth token before running in production.
- Docker Hub pulls over `docker pull` aren't authenticated — if Docker Hub rate-limits the instance, add a login step in `user_data.sh.tftpl`.
