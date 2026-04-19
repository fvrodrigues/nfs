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
  description = "EC2 instance type. Chromium + Go needs a bit of memory; t3.medium is the realistic minimum."
  type        = string
  default     = "t3.medium"
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
  description = "CIDR blocks allowed to reach the app port directly on the EC2 instance. Defaults to the open internet because /prestador has to be reachable by callers without going through API Gateway (API Gateway's 30s timeout is too short for the login+emission flow)."
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

variable "enable_api_gateway" {
  description = "Whether to provision an HTTP API Gateway in front of the instance. The API Gateway is convenient for light endpoints but WILL time out after ~30s on /prestador; prefer hitting the EC2 directly for the main workflow."
  type        = bool
  default     = true
}

variable "api_stage_name" {
  description = "HTTP API Gateway stage. Uses the default stage (auto-deployed) unless you override."
  type        = string
  default     = "$default"
}
