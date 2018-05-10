load("@io_bazel_rules_docker//container:container.bzl", "container_push")

def container_push_internal(name, image, repository):
  container_push(
      name = name,
      format = "Docker",
      image = image,
      registry = "docker-registry.prodrive-technologies.com",
      repository = "it-services-linux/snippets/" + repository,
  )
