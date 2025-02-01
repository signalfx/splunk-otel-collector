#!/bin/bash

# These steps install istio, with a demo profile, set up the default namespace to inject sidecars, CRDs, and the sample application under test.
# See https://istio.io/latest/docs/setup/getting-started/#bookinfo for more info.

curl -L https://istio.io/downloadIstio | sh -
wget https://raw.githubusercontent.com/istio/istio/release-1.24/samples/bookinfo/demo-profile-no-gateways.yaml
$( find . -name istio-*)/bin/istioctl install -f demo-profile-no-gateways.yaml -y
kubectl label namespace default istio-injection=enabled
kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null || \
{ kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v1.2.0" | kubectl apply -f -; }
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.24/samples/bookinfo/platform/kube/bookinfo.yaml
