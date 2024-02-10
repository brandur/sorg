+++
hook = "Costs and benefits of running `golvulncheck` automatically as part of CI."
published_at = 2023-02-20T15:17:59-08:00
title = "Findings from six months of running `govulncheck` in CI"
+++

Last September saw the release of [`govulncheck`](https://go.dev/blog/vuln), a tool that uses that phones into a central database maintained by a Go security team to check for known vulnerabilities that your code might be susceptible to.

It's pretty cool. Instead of doing the easy thing of checking `go.mod` against a list of known modules, it knows specifically which functions are liabilities, and resolves a full call graph of where your code calls into them. It also prints handy links back the vulnerability database along with a succinct summary so you never have to leave your terminal. Here's the output from a [vulnerability in `golang.org/x/net`](https://pkg.go.dev/vuln/GO-2023-1571) last week:

```
Vulnerability #1: GO-2023-1571
  A maliciously crafted HTTP/2 stream could cause excessive CPU
  consumption in the HPACK decoder, sufficient to cause a denial
  of service from a small number of small requests.

  More info: https://pkg.go.dev/vuln/GO-2023-1571

  Module: golang.org/x/net
    Found in: golang.org/x/net@v0.6.0
    Fixed in: golang.org/x/net@v0.7.0

    Call stacks in your code:
Error: client/awsclient/aws_client.go:156:34: awsclient.Client.S3_GetObject
    calls github.com/aws/aws-sdk-go-v2/service/s3.Client.GetObject,
    which eventually calls golang.org/x/net/http2.noDialH2RoundTripper.RoundTrip
```

We've been running it as a CI check since it was released, using a GitHub Actions job:

``` yaml
  vuln_check:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install Go
        uses: actions/setup-go@v5
        with:
          cache: true
          check-latest: true
          go-version: ${{"{{"}} env.GO_VERSION {{"}}"}}

      - name: Install `govulncheck`
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run `govulncheck`
        run: govulncheck ./...
```

## Considerations for CI (#considerations)

The advantages of running `govulncheck` are clear. We find out about problems almost immediately, and usually have fixes deployed within 24 hours of a security notice being made public. Consider for a moment how great that is. Only a few short years ago vulnerabilities in dependencies was almost entirely an unsolved problem and it was overwhelmingly common to be running unpatched code in production _for years_, leading to major compromises of giants like [Equifax](/fragments/gadgets-and-chains) and many others.

But rather than sell this as a straight up always-good-idea, I'll mention the inconveniences as well.

The most common reason the check fails is that a vulnerability is discovered somewhere in Go core and a new patch version of Go like 1.19.6 is released. With the check running in CI, this often leads to a window of a few hours where all our builds are broken because there's a lag between when the version of Go is released and when it lands in GitHub Actions. Luckily GitHub's pretty good about this having [automated the process of upgrading versions](https://github.com/actions/go-versions/commits/main), so it's not bad, but still a little painful.

The worst part of the arrangement is impact on contributors from other teams. When there's a failure, everybody on the team kind of knows what's going on, and we get by until the problem's fixed. But occasionally somebody from a different team will send a change at exactly the wrong time, get hit by it, and be confused. I try to leave an explanatory pull request comment when I'm online, but am often many time zones removed from colleagues, and I'm sure investigating these false positives wastes peoples' time.

It wouldn't be suitable for a larger team. If the builds of ten engineers suddenly start failing simultaneously, it'd be unnecessary chaos and without a clear point person who'd be charge of resolving the situation.

## Out-of-band checkers (#out-of-band)

We're a small company without extensive internal infrastructure. CI's a workable fit for us, and for the time being we don't have anywhere else to run `govulncheck` anyway.

At Stripe this would've lived as part of internal service called "Checker". Checker contained a long list of preconfigured checks that'd run a script or call out to an HTTP endpoint to indicate a pass or fail. Each check has a configured team which owns it, and upon failure it'd open a JIRA ticket assigned to them. An assigned "runner" for the team would look into it, resolve the problem, and the check would flip back to a healthy state.

It was far from a perfect setup. It turns out that it's really easy to write flappy checks and people did in spades. Even with the runner as DRI, failing checks were noisy and tended to bleed into the attention of other team members. "Throw it over the fence" situations were dismally common as people wrote crappy code or burdensome checks because they knew they were handing them off to someone else. Also, JIRA.

But still, an out-of-band system in the spirit of Checker is a better fit for running `govulncheck` than CI is. In a bigger company a security or code platform team would be assigned as owner, and be able to action on a new vulnerability without other contributors even having to know what happened.