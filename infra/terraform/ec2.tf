# Latest Amazon Linux 2023 AMI, resolved at plan time so terraform apply
# always picks a current patched image.
data "aws_ssm_parameter" "al2023_ami" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
}

resource "aws_security_group" "nfse" {
  name        = "${var.project}-sg"
  description = "Inbound rules for the nfse EC2 host"
  vpc_id      = data.aws_vpc.default.id

  # App port (direct hit, bypasses API Gateway). Kept open to the
  # internet so /prestador works without the 30s API Gateway cap.
  ingress {
    description = "nfse HTTP app port"
    from_port   = var.app_port
    to_port     = var.app_port
    protocol    = "tcp"
    cidr_blocks = var.app_ingress_cidrs
  }

  # Optional SSH. Leave var.ssh_ingress_cidrs empty to disable.
  dynamic "ingress" {
    for_each = length(var.ssh_ingress_cidrs) > 0 ? [1] : []
    content {
      description = "SSH"
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = var.ssh_ingress_cidrs
    }
  }

  egress {
    description = "all egress (docker pull, portal access, 2Captcha)"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-sg"
  }
}

# IAM role so the host can be managed over SSM Session Manager without
# opening SSH. Lets you `aws ssm start-session` into the box even when
# ssh_ingress_cidrs is empty.
resource "aws_iam_role" "nfse" {
  name = "${var.project}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.nfse.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "nfse" {
  name = "${var.project}-ec2-profile"
  role = aws_iam_role.nfse.name
}

resource "aws_instance" "nfse" {
  ami                         = data.aws_ssm_parameter.al2023_ami.value
  instance_type               = var.instance_type
  subnet_id                   = data.aws_subnets.default.ids[0]
  vpc_security_group_ids      = [aws_security_group.nfse.id]
  iam_instance_profile        = aws_iam_instance_profile.nfse.name
  associate_public_ip_address = var.associate_public_ip
  key_name                    = var.key_pair_name != "" ? var.key_pair_name : null

  user_data = templatefile("${path.module}/user_data.sh.tftpl", {
    docker_image   = var.docker_image
    container_name = var.project
    app_port       = var.app_port
  })

  # user_data changes should force a replacement so the container gets
  # re-pulled and restarted cleanly. If you'd rather update in place,
  # drop this block and run `terraform taint aws_instance.nfse` instead.
  user_data_replace_on_change = true

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    encrypted   = true
  }

  metadata_options {
    http_tokens   = "required"
    http_endpoint = "enabled"
  }

  tags = {
    Name = var.project
  }
}
