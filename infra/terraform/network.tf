# Use the account's default VPC/subnets to keep this example minimal. If
# you need to run this in a dedicated VPC, swap the data sources for an
# explicit aws_vpc / aws_subnet block or switch to the
# terraform-aws-modules/vpc module.

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}
