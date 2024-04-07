This folder contains the definition of various Docker images used in testing.

Those images are stored under quay.io/splunko11ytest, with the image name matching the folder name.

When testing locally,
- Build and start individual service(s):
  ```
  docker compose -f docker/docker-compose.yml up --build --detach <service1> <service2> ...
  ```
- Build and start all services for non smart agent tests:
  ```
  docker compose -f docker/docker-compose.yml --profile integration up --build --detach
  ```
- Build and start all services for smart agent tests:
  ```
  docker compose -f docker/docker-compose.yml --profile smartagent up --build --detach
  ```

When adding/modifying service images, ensure the directory name under [docker](../docker) matches the image name in
[docker-compose.yml](./docker-compose.yml).

If you need to update an image, update the Dockerfile with a PR. After merge, push the image manually with:

```
cd docker/<image>
docker buildx build --platform=linux/amd64 --push -t quay.io/splunko11ytest/<image>:latest .
```
