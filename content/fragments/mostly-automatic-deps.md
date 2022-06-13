+++
hook = "Writing a GitHub Action to bump Go depenendencies and put the change in a pull request."
published_at = 2022-06-07T16:12:27Z
title = "Mostly automatic Go dependency updates with GitHub Actions"
+++

**Update (2022/06/12):** [Michael points out](https://twitter.com/embano1/status/1535871515941642241) that you can probably just GitHub's [built-in Dependabot](https://github.blog/2020-06-01-keep-all-your-packages-up-to-date-with-dependabot/) to do this as well.

An issue that came up during a [recent pen test](/fragments/crypto-rand-float64) on our services is that we had quite a few outdated packages deteriorating in our Go stack. Not super surprising given that there's no function, either programmatic or process-wise, that ever causes any of our dependencies to be updated, except when we do so manually on rare occasions to get a bug fix or new feature from one of them.

Go Modules makes dependency updates quite easy. This command will bump everything where a new patch or minor version is found (the `-t` also includes test packages):

``` sh
$ go get -t -u ./...
```

Usually you also want to couple this with a tidy to make sure that everything that doesn't need to be there gets squeezed back out:

``` sh
$ go mod tidy
```

Even so, someone still has to remember to run those commands. So that's where GitHub Actions, once again showing its tremendous utility and flexibility, comes to the rescue. This week we added a job that triggers off cron every Monday to bump all our dependencies, run tidy, and open a pull request with the result, requesting review from someone on the team:

``` yaml
{{HTMLSafePassThrough `
env:
  GO_VERSION: 1.18

on:
  workflow_dispatch:
  schedule:
    - cron: "0 17 * * 1" # 10am pdt / 9am pst, weekly on Monday

jobs:
  dep_update:
    runs-on: ubuntu-latest
    timeout-minutes: 10

    steps:
      - name: Install Go
        uses: actions/setup-go@v2
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Checkout
        uses: actions/checkout@v3

      - name: Update minor and patch-level dependencies
        run: go get -t -u ./...

      - name: Tidy
        run: go mod tidy

      - name: Create pull request
        uses: peter-evans/create-pull-request@v4
        with:
          title: "Update package dependencies + tidy"
          body: |
            This is a change initiated automatically on a weekly basis by a
            GitHub Action that updates the projects dependencies to their latest
            minor and patch-level versions. This lets us stay up to date
            incrementally so that updates are less effort to get merged compared
            to large monolithic updates, and gets us security updates more
            expediently.

            If the build passes, you are probably A-OK to merge and deploy this.
            If not, try to dig into what's not working and see if you can fix it
            so that the dep train stays on its rails.

            Note that although minor/patch level changes are handled
            automatically, notably major version changes like you'd find in
            stripe-go are not and those upgrades need to be performed manually.
            That should theoretically not be a problem if fixes are backported
            to all previous majors, but in practice they are often not, so it's
            worthwhile to occasionally look for new majors and integrate them.
          branch: "dep-update"
          commit-message: |
            Update package dependencies + tidy

            Weekly update to the project's package dependencies initiated by an
            automatic GitHub Action running on cron. Keeps upgrades less of a
            monolithic task and lets security-related patches trickle in more
            quickly.
          author: "Bot <bot@crunchydata.com>"
          committer: "Bot <bot@crunchydata.com>"
          delete-branch: true
          reviewers: |
            brandur
`}}
```

There's still some manual work in that a human needs to look at what's in there and do the merge, but this is preferable to the alternative where a merge happens automatically, but breaks something. Our Ruby team has been using a similar process for months now, and it's working well for them -- not being too much of a burden for someone to take a quick glance at the PR at the beginning of every week. I'm hoping we find similar luck.

An edge not handled is that we don't try to do anything about new _major_ versions of dependencies. This should theoretically be okay because bug and security fixes are backported, but in many cases this doesn't actually happen (I can tell you from first-hand experience that it doesn't with stripe-go for example). This isn't a huge problem because for the majority of packages in Go new majors are rare, but it'd be nice to bake in some process/automation that gets us to look at this every so often.
