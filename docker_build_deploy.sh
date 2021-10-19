########### Configure Container Names ###########

VERSION=0.31.0

MULTI_ARCH_IMAGE_NAME=signalfx/otelcol:$VERSION-m-shell
AMD_IMAGE_NAME=signalfx/otelcol:$VERSION-amd64-shell
ARM_IMAGE_NAME=signalfx/otelcol:$VERSION-arm64-shell

SMART_AGENT_RELEASE=v5.11.2

#################################################

# Copy AMD binary into diretory 
cp ./bin/otelcol_linux_amd64 ./cmd/otelcol/otelcol
# Build AMD container 
docker build -t $AMD_IMAGE_NAME --network host --build-arg SMART_AGENT_RELEASE=$SMART_AGENT_RELEASE ./cmd/otelcol/
docker push $AMD_IMAGE_NAME
# Clean 
rm ./cmd/otelcol/otelcol

# Copy ARM binary into diretory 
cp ./bin/otelcol_linux_arm64 ./cmd/otelcol/otelcol
# Build ARM64 container 
docker build -t $ARM_IMAGE_NAME --network host --build-arg SMART_AGENT_RELEASE=$SMART_AGENT_RELEASE ./cmd/otelcol/
docker push $ARM_IMAGE_NAME
rm ./cmd/otelcol/otelcol

# Create a new multi-arch manifest manifest 
docker manifest create --amend $MULTI_ARCH_IMAGE_NAME $AMD_IMAGE_NAME $ARM_IMAGE_NAME

# Annotate the manifest layers for arm64
docker manifest annotate $MULTI_ARCH_IMAGE_NAME $ARM_IMAGE_NAME --os linux --arch arm64
docker manifest push $MULTI_ARCH_IMAGE_NAME 