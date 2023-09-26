+++
hook = "Get occasionally-indispensible Postgres logging detail with three lines of YAML."
published_at = 2023-09-26T15:44:02-07:00
title = "Getting Postgres logs in a GitHub Action"
+++

I have an admission to make: I don't read Postgres logs all that often. The majority of the time I get enough feedback in-band with Postgres operations that I don't need to look at them.

But occasionally, they're absolutely indispensible. Recently I've been debugging an upsert deadlock problem in our test suite, and the `deadlock detected (SQLSTATE 40P01)` error message that comes back with the failed operation is entirely unactionable in every respect. The logs on the other hand, contain a rich vein of information, including the full queries and PIDs that conflicted with each other.

I was having trouble reproducing this particular problem locally, having to rely entirely on CI to diagnosis it, but was having trouble doing so without access to Postgres logs there.

Typically, Postgres is run in GitHub Actions as a service container:

``` yaml
services:
  postgres:
    image: postgres
    env:
      POSTGRES_PASSWORD: postgres
    ports:
      # Maps tcp port 5432 on service container to the host
      - 5432:5432
```

The special `postgres` image tag maps to the [Docker Hub Postgres image](https://hub.docker.com/_/postgres), which is customizable to some extent using environmental variables like `POSTGRES_PASSWORD`.

Postgres logs to stderr by default. It can be reconfigured to log to file or syslog with [`log_destination`](https://www.postgresql.org/docs/current/runtime-config-logging.html), but the Docker image provides only limited knobs for extensibility, and doing so would involve writing a post-processing script to bring in a customized Postgres conf file. Ugly, and also a lot of work.

## Copy/paste starting here (#copy-paste)

Luckily, there's a dead simple alternative. An invocation of `docker logs` returns lines emitted to stdout/stderr, and is compatible with GitHub Actions service containers. Add this step **near the end** of a job, after interesting things like tests have already run:

``` yaml
{{HTMLSafePassThrough `
- name: Postgres logs
  if: always() # run even on failure
  run: docker logs "${{ job.services.postgres.id }}"
`}}
```

Details:

* We use `if: always()` to tell GitHub that the step should be run unconditionally, even in cases where the preceeding step failed. So if the tests fail, Postgres logging is available to help explain why.

* The special `{{HTMLSafePassThrough `${{ job.services.postgres.id }}`}}` tag is used to feed a container ID to `docker logs`. Service containers are addressed according to their name. Ours is called simply `postgres` (see the first block of YAML above), but this value may be different depending on the naming in YAML.

* Postgres' default logging level is `NOTICE`, which is above `INFO` and below `WARN`. It produces extra detail when errors occur, but doesn't overwhelm `docker logs` with a sea of low fidelity debugging output, making it perfect for this purpose.