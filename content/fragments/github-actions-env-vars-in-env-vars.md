+++
hook = "Using environment files in GitHub Actions to define variables that can use values of other variables."
published_at = 2020-12-03T23:46:05Z
title = "GitHub Actions: Setting env vars with other env vars"
+++

GitHub Actions provides an `env` configuration that allows a build to inject environmental variables into a full workflow or an individual step:

``` yaml
jobs:
  build:
    env:
      PG_DATA_DIR: /home/runner/data
```

String literals are fine, but users will find out the hard way that using environment variables as inputs to other environment variables **does not work**:

``` yaml
jobs:
  build:
    env:
      PG_DATA_DIR: $HOME/data
```

The reason is that the values are slurped up when a workflow's YAML is being parsed, and never interpreted through a shell that would enable variable expansion. `$HOME/data` above comes out as _literally_ `$HOME/data` instead of the intended `/home/runner/data`.

## The fix (#fix)

The workaround is to use GitHub Actions [environment files](https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#environment-files) [1]. Values written to `$GITHUB_ENV` are available in subsequent steps:

``` yaml
jobs:
  build:
    steps:
      - name: "Set environmental variables"
        run: |
          echo "PG_DATA_DIR=$HOME/data" >> $GITHUB_ENV

      - name: "Can use environment variables"
        run: |
          echo "Working variable from variable: $PG_DATA_DIR"
```

[1] It was previously possible to use `set-env` to do the same thing, but that command [has been deprecated](https://github.blog/changelog/2020-10-01-github-actions-deprecating-set-env-and-add-path-commands/) due to a [security flaw](https://bugs.chromium.org/p/project-zero/issues/detail?id=2070) discovered in it by Google's Project Zero.
