variable "account_id" {
  type        = string
  description = "The AWS account ID where resources will be deployed."
}

variable "kubernetes_namespace" {
  type        = string
  description = "The Kubernetes namespace for the service account trust relationship."
  default     = "enterprisearchitecture-pnpbj"
}

variable "eks_oidc_provider_ids" {
  type        = list(string)
  description = "OIDC provider IDs for EKS clusters (Hellman production and diffe standby)."
  default = [
    "B182759F93D251942CB146063F57036B",
    "2A3537C379F58D1212A72BD93332F5C9",
  ]
}

variable "kms_keys" {
  type        = list(string)
  description = "A list of KMS key IDs to be accessed."
  default     = ["*"]
}

variable "secretsmanager_secret_names" {
  type        = list(string)
  description = "A list of Secrets Manager secret names to be accessed."
  default = [
    "easi/db-credentials-*",
    "easi/oidc-credentials-*",
    "easi/platform-credentials-*",
    "easi/encryption-key-*",
    "easi/agent-token-secret-*",
    "rds!db-30146832-1030-47da-839f-d3ffaf6b2e33-*",
  ]
}
