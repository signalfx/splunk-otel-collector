This folder contains the definition of various Docker images used in testing.

Those images are stored under quay.io/splunko11ytest, with the image name matching the folder name.

If you need to update an image, update the Dockerfile with a PR. After merge, push the image manually with:

```
cd docker/<image>
docker buildx build --platform=linux/amd64 --push -t quay.io/splunko11ytest/<image>:latest .
```
