# Production Deployment Pipeline

## Description
Implement production deployment pipeline for Azure DevOps.
Use AWS RDS for postgres db, following this template: https://github.com/dfds/infrastructure-modules/tree/master/database/postgres 
The backend and frontend must run containerised on k8s. There's a sample pipeline here: https://github.com/dfds/selfservice-portal/blob/develop/azure-pipelines.yml

## Requirements
- Azure DevOps pipeline configuration with build and deploy stages
- Production Dockerfiles for backend (Go) and frontend (React/nginx)
- Kubernetes manifests for deployments, services, and ingress
- AWS RDS PostgreSQL infrastructure using DFDS Terraform module
- Database credentials management via Kubernetes secrets
- Automated testing in pipeline
- Short deployment documentation

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Documentation updated if needed
- [x] User sign-off
