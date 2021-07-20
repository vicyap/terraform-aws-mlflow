variable "random_id" {
  type = string
}

variable "is_private" {
  type = bool
}

variable "artifact_bucket_id" {
  default = null
}

variable "service_image" {
  default = "larribas/mlflow"
}

variable "service_image_tag" {
  default = "1.9.1"
}

## begin networking variables

variable "vpc_id" {
  type = string
}

variable "vpc_cidr_block" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "public_subnet_ids" {
  type = list(string)
}

variable "database_subnet_ids" {
  type = list(string)
}

## end networking variables
