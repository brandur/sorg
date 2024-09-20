+++
hook = "How I accidentally stumbled across the `source_env` directive and dramatically improved my configuration methodology overnight."
published_at = 2024-09-20T04:53:58-07:00
title = "Direnv's `source_env`, and how to manage project configuration"
+++

For years I've been using [Direnv](https://direnv.net/) to manage configuration in projects. It's a small program that loads env vars out of an `.envrc` file on a directory by directory basis, using a shell hook to load vars as you enter a folder, and unload them as you leave.

A typical `.envrc`:

``` sh
export API_URL="http://localhost:5222"
export DATABASE_URL="postgres://localhost:5432/project-db"
export ENV_NAME=dev
```

The beauty of Direnv is not only that it's 12-factor friendly, but that it's language agnostic, and unlike its language-specific alternatives that hook into program code in various creative ways, Direnv makes configuration available to your main program _and_ anything else you need to run with it.

So configuration is available for your project's core programs:

``` sh
# gets DATABASE_URL from env
make build/api && build/api
```

And for all adjacent utilities, including ones that you didn't write, and would otherwise have no way of hooking into a bespoke configuration system:

``` sh
# still works fine!
goose -dir ./migrations/main postgres $DATABASE_URL
```

## Uneven distribution (#uneven-distribution)

For years I've recommended in project READMEs to get started by copying an `.envrc` template and running the program:

``` sh
cp .envrc.sample .envrc
direnv allow
go test ./...
```

`.envrc.sample` is committed to Git while `.envrc` is not due to the presumption that it may eventually be edited to include user-specific secrets.

That works fine, but has always had the downside in that if configuration changes and `.envrc.sample` is updated, other developers don't get those changes unless they copy a fresh `.envrc.sample`, and they almost certainly won't think to do that. This is an advantage that I'd thought language-specific configuration systems like [Dotenv](https://www.npmjs.com/package/dotenv0) have had over Direnv, where they can often read multiple env files, some of which may contain shared configuration that's versioned with the repo.

## The missing piece of the puzzle: `source_env`

Well, after being a Direnv user for _ten years_, yesterday I learnt of the existence of [`source_env`](https://direnv.net/man/direnv-stdlib.1.html), a special directive that can go in an `.envrc` and which will read out out of another envrc file.

This simplifies the configuration of my projects _dramatically_. They have an `.envrc.sample`, but it's stripped down to almost nothing, containing only a `source_env` statement and room to add customization.

``` sh
# Common configuration for al developers, committed to Git.
source_env .envrc.local

# Custom env values go here.
```

Meanwhile, all default configuration migrates to a `.envrc.local` (the `.local` suffix not having any special meaning, but rather just a convention to use):

``` sh
#
# .envrc.local
#
# Shared env vars commmitted to Git and made available to all
# developers. As # much configuration should go here as possible
# so that new env vars don't break # anyone and everyone gets to
# benefit from improvements, but don't add anything too secret or
# too custom.
#

export API_URL="http://localhost:5222"
export DATABASE_URL="postgres://localhost:5432/project-db"
export ENV_NAME=dev
```

`.envrc.local` is committed to Git, and when anyone changes configuration, all other developers get the updates the next time they pull from master.

This doesn't account for truly sensitive configuration that shouldn't be stored in a Git repository, but my advice on that: projects should always be able to gracefully degrade so they can run (at least in development mode) with no sensitive secrets at all. And _certainly_ the test suite should be able to. If your project can't do that, something is wrong.

For my money, Direnv + `source_env` is a perfect dev configuration system, and one that works cleanly in any language ecosystem.