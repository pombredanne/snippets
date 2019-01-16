# Snippets

<p align="center">
  <img src="https://github.com/ProdriveTechnologies/snippets/raw/master/screenshot.png" alt="Screenshot of the Snippets web application" width="525" height="416"/>
</p>

Snippets is a small web application that can be used to track diary
entries on a weekly basis. By allowing users of this web application to
subscribe to others, they may receive weekly reports of all snippets
written that week. Work on this utility was inspired by
[an identically named tool in use at Google](https://www.inc.com/jessica-stillman/a-simple-productivity-tip-from-googles-early-days.html).

# Running Snippets

1. Create your own [Bazel](https://bazel.build/) workspace that contains
   Snippets as a dependency (either as `http_archive`, `git_repository`
   or `go_repository`), e.g.:
   ```python
   http_archive(
       name = "snippets",
       sha256 = "<checksum of source tarball>",
       strip_prefix = "snippets-<tag or commit>",
       url = "https://github.com/ProdriveTechnologies/snippets/archive/<tag or commit>.tar.gz",
   )
   ```
   You will also need to pull in Go dependencies from Snippets, which you
   can find in its [WORKSPACE](https://github.com/ProdriveTechnologies/snippets/blob/master/WORKSPACE).
   In your BUILD.bazel, add `container_push()` directives to push Docker
   containers of the individual applications into your own Docker registry,
   e.g.:
   ```python
   container_push(
       name = "snippets_cron_reminders_push",
       format = "Docker",
       image = "@snippets//cmd/snippets_cron_reminders:snippets_cron_reminders_container",
       registry = "my-docker-registry.com",
       repository = "snippets_cron_reminders",
   )
   ```
1. Run the following commands in your Bazel workspace to push the
   container images:
   ```sh
   bazel build //...
   for i in $(bazel query //... | grep '_push$'); do
       bazel run $i
   done
   ```
1. Create a PostgreSQL or [CockroachDB](https://www.cockroachlabs.com/)
   database with the tables specified in `database_schema.sql`.
1. Run the `snippets_web` container to enable the Snippets web application.
   Place an authenticating proxy, such as
   [keycloak-proxy](https://github.com/gambol99/keycloak-proxy) in front of it
   that at least sets the headers `X-Auth-Subject`, `X-Auth-Name` and
   `X-Auth-Email`, containing the user's username, real name and email
   address, respectively.
1. Set up a cronjob that runs the `snippets_cron_reminders` container on
   Fridays to send weekly reminders to users of the service, so that
   they don't forget to write a snippet.
1. Set up a cronjob that runs the `snippets_cron_subscriptions`
   container on Mondays to send copies of snippets written in the
   previous week to subscribers.

Each of the containers can be configured by providing command line
flags. Please refer to the `main.go` source files or start the
containers with `-help` to get a list of supported command line flags.

# Background

Snippets has been written by @EdSchouten and @mickael-carl for use at
Prodrive Technologies B.V. This repository is intended to act as a
canonical example of how a Go-based application may be structured.
