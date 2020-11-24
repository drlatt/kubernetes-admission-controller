# Kubernetes Admission Controller

## This repo contains an admission controller which accepts or rejects requests to create pods based on rules specified in ![the instructions file](./INSTRUCTIONS.md)

# Usage
To run the code in this repo, you need to ensure you have the `make utility`, `kubectl`, `docker` and `minikube` installed on your PC.

## Steps to run
1. Run `make` to get a help message on helpful make commands to run the code.
```
Usage: make <target>
  mod         download modules to local cache
  docker      build and tag docker image
  push        push docker image to registry
  deploy      deploy webhook server to kubernetes cluster
  deployad    deploy admission components to kubernetes cluster
  clean       remove deployed components from kubernetes cluster and delete created images
  app         create docker image and deploy kubernetes components
  local       build application binary locally
```
2. Create ssl certificates for the application service using the script `webhook-create-signed-cert.sh`. Store these certs in the `ssl-certs` folder created in this repo.
3. Obtain the certificate-authority-data from your kubernetes cluster using this command: `kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}'`. Once obtained, use it to replace the CA_BUNDLE variable in the `webhook-server.yaml` file.
4. Run `make deploy` to run the application. This pulls the pre-built image from dockerhub and deploys it to your kubernetes cluster by applying the `webhook-server.yaml` file. This file contains the specifications for the admission controller, the application deployment and the associated service.
5. Run `make clean` to remove components created by the `make deploy` command.
 
## N.B. To build your own image and deploy, you can run `make app`. However, remember to change the IMAGE_REPO_NAME variable in the Makefile. This ensures the resulting image is tagged and pushed to your personal container repo.

## Bug/improvement: Deployments without the `team` label are accepted to the cluster but the patch isn't made to add the `team` label to the pod's set of labels.
