resource "aws_secretsmanager_secret" "db_password" {
  name_prefix = "mlflow-terratest"
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = "ran${var.random_id}dom"
}

module "mlflow" {
  source = "../../../"

  unique_name = "mlflow-terratest-${var.random_id}"
  tags = {
    "owner" = "terratest"
  }
  vpc_id                            = var.vpc_id
  database_subnet_ids               = var.database_subnet_ids
  service_subnet_ids                = var.private_subnet_ids
  load_balancer_subnet_ids          = var.is_private ? var.private_subnet_ids : var.public_subnet_ids
  load_balancer_ingress_cidr_blocks = var.is_private ? [var.vpc_cidr_block] : ["0.0.0.0/0"]
  load_balancer_is_internal         = var.is_private
  artifact_bucket_id                = var.artifact_bucket_id
  database_password_secret_arn      = aws_secretsmanager_secret_version.db_password.secret_id
  database_skip_final_snapshot      = true
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = module.mlflow.load_balancer_arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = module.mlflow.load_balancer_target_group_id
    type             = "forward"
  }
}
