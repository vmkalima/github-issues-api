#!/usr/bin/env bash
set -euo pipefail

echo "** Checking minikube status..."
if ! minikube status > /dev/null 2>&1; then
	echo "** minikube not running, starting it..."
	minikube start --driver=docker
fi

echo "** Pointing Docker at minikube's daemon..."
eval $(minikube docker-env)

echo "** Building image inside minikube..."
docker build -t github-issues-api:latest .

echo "** Checking for secret..."
if [ ! -f deploy/k8s/secret.yaml ]; then
	echo "ERROR: deploy/k8s/secret.yaml not found. Copy deploy/k8s/secret.example.yaml and set a real value first."
	exit 1
fi

echo "** Applying manifests..."
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml

echo "** Restarting deployment to pick up new image..."
kubectl rollout restart deployment github-issues-api
kubectl rollout status deployment/github-issues-api --timeout=90s

kubectl port-forward svc/github-issues-api-service 8080:80
echo "** Deployment ready. Use another terminal to make requests to the API."