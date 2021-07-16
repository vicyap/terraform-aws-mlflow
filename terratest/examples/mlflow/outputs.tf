# Outputs for Terratest to use
output "load_balancer_dns_name" {
  value = module.mlflow.load_balancer_dns_name
}

output "artifact_bucket_id" {
  value = module.mlflow.artifact_bucket_id
}
