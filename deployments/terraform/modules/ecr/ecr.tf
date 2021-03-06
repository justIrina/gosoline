resource "aws_ecr_repository" "main" {
  name = "${var.project}/${var.family}/${var.application}"
}

resource "aws_ecr_lifecycle_policy" "main" {
  repository = aws_ecr_repository.main.name
  count      = var.use_default_lifecycle_policy

  policy = <<EOF
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Keep only 10 master images",
            "selection": {
                "tagStatus": "tagged",
                "tagPrefixList": ["${var.primary_tag}"],
                "countType": "imageCountMoreThan",
                "countNumber": 10
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 2,
            "description": "Expire any image older than 7 days",
            "selection": {
                "tagStatus": "any",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 7
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}
