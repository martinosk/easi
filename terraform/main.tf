terraform {
  backend "s3" {
  }
}

provider "aws" {
  region = "eu-west-1"
}

module "db_instance" {
  source = "git::https://github.com/dfds/terraform-aws-rds.git?ref=2.4.0"

  #     Provide a cost centre for the resource.
  #     Valid Values: .
  #     Notes: This set the dfds.cost_centre tag. See recommendations [here](https://wiki.dfds.cloud/en/playbooks/standards/tagging_policy).
  cost_centre = "ti-arch"

  #     Specify data classification.
  #     Valid Values: public, private, confidential, restricted
  #     Notes: This set the dfds.data.classification tag. See recommendations [here](https://wiki.dfds.cloud/en/playbooks/standards/tagging_policy).
  data_classification = "confidential"

  #     Specify the staging environment.
  #     Valid Values: "dev", "test", "staging", "uat", "training", "prod".
  #     Notes: The value will set configuration defaults according to DFDS policies.
  environment = "prod"

  #     Specify the name of the RDS instance to create.
  #     Valid Values: .
  #     Notes: .
  identifier = "easi"

  #     [Experiemental Feature] Specify whether or not to deploy the instance as multi-az database cluster.
  #     Valid Values: .
  #     Notes:
  #     - This feature is currently in beta and is subject to change.
  #     - It creates a DB cluster with a primary DB instance and two readable standby DB instances,
  #     - Each DB instance in a different Availability Zone (AZ).
  #     - Provides high availability, data redundancy and increases capacity to serve read workloads
  #     - Proxy is not supported for cluster instances.
  #     - For smaller workloads we recommend considering using a single instance instead of a cluster.
  is_cluster = false

  #     Specify whether or not to enable access from Kubernetes pods.
  #     Valid Values: .
  #     Notes: Enabling this will create the following resources:
  #       - IAM role for service account (IRSA)
  #       - IAM policy for service account (IRSA)
  #       - Peering connection from EKS Cluster requires a VPC peering deployed in the AWS account.
  is_kubernetes_app_enabled = true

  #     Specify whether or not to include proxy.
  #     Valid Values: .
  #     Notes: Proxy helps managing database connections. See [documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy-planning.html) for more information.
  is_proxy_included = false

  #     Specify whether or not this instance is publicly accessible.
  #     Valid Values: .
  #     Notes:
  #     - Setting this to true will do the followings:
  #       - Assign a public IP address and the host name of the DB instance will resolve to the public IP address.
  #       - Access from within the VPC can be achived by using the private IP address of the assigned Network Interface.
  #       - Create a security group rule to allow inbound traffic from the specified CIDR blocks.
  #         - It is required to set `public_access_ip_whitelist` to allow access from specific IP addresses.
  is_publicly_accessible = false

  #     Specify service availability.
  #     Valid Values: low, medium, high
  #     Notes: This set the dfds.service.availability tag. See recommendations [here](https://wiki.dfds.cloud/en/playbooks/standards/tagging_policy).
  service_availability = "low"

  #     Provide a list of VPC subnet IDs.
  #     Valid Values: .
  #     Notes: IDs of the subnets must be in the same VPC as the RDS instance. Example: ["subnet-aaaaaaaaaaa", "subnet-bbbbbbbbbbb", "subnet-cccccccccc"]
  subnet_ids = ["subnet-08f10b6cb27fad938", "subnet-0b4e07915c8a91427", "subnet-08ab2e2ff52cf8ad0"]

  #     Specify Username for the master DB user.
  #     Valid Values: .
  #     Notes: .
  username = "easidbuser"

  #     Specify the VPC ID.
  #     Valid Values: .
  #     Notes: .
  vpc_id = "vpc-02d14cf9eae0f7280"
}

output "iam_instance_profile_for_ec2" {
  description = "The name of the EC2 instance profile that is using the IAM Role that give AWS services access to the RDS instance and Secrets Manager"
  value       = try(module.db_instance.iam_instance_profile_for_ec2, null)
}
output "iam_role_arn_for_aws_services" {
  description = "The ARN of the IAM Role that give AWS services access to the RDS instance and Secrets Manager"
  value       = try(module.db_instance.iam_role_arn_for_aws_services, null)
}
output "kubernetes_serviceaccount" {
  description = "If you create this Kubernetes ServiceAccount, you will get access to the RDS through IRSA"
  value       = try(module.db_instance.kubernetes_serviceaccount, null)
}
output "peering" {
  description = "None"
  value       = try(module.db_instance.peering, null)
}