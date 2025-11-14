remote_state {
  backend = "s3"
  config = {
    bucket         = "easi-state-bucket"
    encrypt        = true
    key            = "easi-prod/terraform.tfstate" # This is the path to the state file inside the bucket.
    region         = "eu-west-1"
    dynamodb_table = "terraform-locks"
  }
}