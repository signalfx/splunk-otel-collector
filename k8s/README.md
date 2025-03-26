This folder contains a set of scripts that execute in separate runs to create Kubernetes environments where a specific
integration test will run to test discovery.

The folder name (`envoy`, `istio`) will match the equivalent make target 
(`make integration-test-envoy-discovery-k8s`, `make integration-test-istio-discovery-k8s`).

Reference: `.github/workflows/integration-test.yaml`.
