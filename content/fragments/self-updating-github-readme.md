+++
hook = "Building a GitHub profile `README.md` that updates itself automatically using Go and GitHub Actions."
published_at = 2020-08-29T20:41:07Z
title = "Building a self-updating GitHub README"
+++

I stole [Simon Willison's idea](https://simonwillison.net/2020/Jul/10/self-updating-profile-readme/) and built a self-updating GitHub `README.md` profile:

![Screenshot of GitHub profile page](/assets/fragments/self-updating-github-readme/github-profile-readme.png)

It's driven by a [tiny Go program](https://github.com/brandur/brandur/blob/master/updater/main.go) that's run periodically in a scheduled GitHub Action. It pulls down a couple Atom feeds, renders a template, and pushes the result if there was a change.

I expected that getting the right credentials in place to do a Git push from an Actions job would be painful, and was pleased to learn that it requires no configuration whatsoever -- version 2 of the checkout action (`actions/checkout@v2`) streamlines the process totally. Here's the critical step in my [workflow file](https://github.com/brandur/brandur/blob/master/.github/workflows/ci.yml):

``` yaml
- name: Commit and push (if changed)
  run: |-
    git diff
    git config --global user.email "actions@users.noreply.github.com"
    git config --global user.name "README-bot"
    git add -u
    git commit -m "Automatic update from GitHub Action" || exit 0
    git push
  if: github.ref == 'refs/heads/master'
```

Note the  `|| exit 0` that bails cleanly if there wasn't a change.

The project doesn't have huge technical merit, but it was just ... fun. There's something really satisfying about connecting tiny programs together that will run on maintenance-free serverless infrastructure. Go's tooling and syntax has shown itself to be [so stable](https://brandur.org/10000-years) that I expect this automation to be able to run for years without me having to fix anything. GitHub's choice of Markdown as the format for profile pages is an ideal choice for creative engineers in that it enables stylistic self-expression, but with tight constraints. You've got all your basic HTML elements and emoji, but arbitrary HTML won't render.
