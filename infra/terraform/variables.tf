variable "project" {
  description = "Name used as prefix for created resources and as the default tag."
  type        = string
  default     = "nfse"
}

variable "region" {
  description = "AWS region to deploy the infrastructure into."
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type. Defaulted to t2.micro for sandbox/free-tier compatibility. WARNING: t2.micro (1 GB RAM) is too small to actually run the Chromium-backed /prestador flow; the container will OOM on the first request. For a functional deployment, use t3.medium or larger."
  type        = string
  default     = "t2.micro"
}

variable "docker_image" {
  description = "Container image the EC2 host should pull and run."
  type        = string
  default     = "rodriguesflavio/nfse:latest"
}

variable "app_port" {
  description = "Port the nfse container listens on (container and host side are kept equal)."
  type        = number
  default     = 8080
}

variable "app_ingress_cidrs" {
  description = "CIDR blocks allowed to reach the app port directly on the EC2 instance. Tighten to your callers' CIDR in production."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "ssh_ingress_cidrs" {
  description = "CIDR blocks allowed to SSH into the EC2 instance. Leave empty to disable SSH entirely (recommended in production)."
  type        = list(string)
  default     = []
}

variable "key_pair_name" {
  description = "Optional name of an existing EC2 key pair. If empty, no key pair is attached and SSH is unavailable."
  type        = string
  default     = ""
}

variable "associate_public_ip" {
  description = "Whether to associate a public IP with the EC2 instance. Required if you want to reach :8080 from the internet directly."
  type        = bool
  default     = true
}
