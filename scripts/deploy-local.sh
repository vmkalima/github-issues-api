#!/usr/bin/env bash
set -euo pipefail

echo "** Checking minikube status..."
if ! minikube status > /dev/null 2>&1; then
    echo "** minikube not running, starting it..."
    minikube start --driver=docker
fi

echo "** Pointing Docker at minikube's daemon..."
eval "$(minikube docker-env)"

echo "** Building image inside minikube..."
docker build -t github-issues-api:latest .

echo "** Applying manifests..."
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml

echo "** Restarting deployment to pick up new image and secrets..."
kubectl rollout restart deployment github-issues-api

echo "** Waiting for deployment..."
kubectl rollout status deployment/github-issues-api --timeout=90s

echo "** Deployment ready."
echo "** Forwarding localhost:8080 -> Kubernetes service:80"
echo "** Press Ctrl+C to stop port-forwarding."

kubectl port-forward svc/github-issues-api-service 8080:80