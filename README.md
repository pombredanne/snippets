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

1. Create your own [Bazel](https://bazel.build/) workspace similar to
   the one provided in the `example-workspace/` directory that contains
   `container_push()` directives to push Docker containers of the
   individual applications into your own Docker registry.
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

Snippets has been written by Ed Schouten and Mickaël Carl for use at
Prodrive Technologies B.V. This repository is intended to act as a
canonical example of how a Go-based application may be structured.
