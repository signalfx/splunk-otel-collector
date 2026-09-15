This folder contains the definition of various Docker images used in testing.

These images are built locally from their Dockerfiles. Integration CI builds the images for each test profile and does
not use the published test images or a remote build cache.

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

If you need to update an image, update its Dockerfile with a PR. The image will be rebuilt by CI when the corresponding
test profile runs.
