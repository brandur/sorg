+++
hook = "Automated code deployments to Google Cloud Run with only a few dozen lines of Actions configuration."
published_at = 2020-04-06T04:22:23Z
title = "Development log: Deploying Google Cloud Run from GitHub Actions"
+++

I've been putting one foot in the water recently in testing [Google Cloud Run](https://cloud.google.com/run/docs) as a Heroku alternative for hobby projects. It's got a somewhat scary pricing model that gets expensive if you have a program serving sustained traffic around the clock, but the billing is per-second, and in practice should cost very little for many apps. I've had a few basic ones running on it for a few months now, and so far only overshot the free tier by ~20 cents.

The rest of the product seems well-built. Unlike other forms of serverless like AWS Lambda, running containers are allowed to serve more than a one request at a time, making deployments countless times faster and more efficient compared to the alternative containers-as-concurrency model. Deploys are all based on the [OCI](https://github.com/opencontainers/image-spec) (Open Container Initiative) image format, gradual rollouts are supported, it's easy getting automated TLS for a custom domain set up, and scaling is fast, easy, and automatic.

## Automating ships (#automating-ships)

A step I took today was having my project deploy automatically from GitHub Actions. Google Cloud is commonly manipulated using `gcloud`, which is a CLI alternative to using their web console. For GitHub Actions, Google's taken an intuitive and pragmatic approach which involves a single setup step for getting `gcloud` configured and in place, then using it to run operations the same way as from a local box.

The [`setup-gcloud` action](https://github.com/GoogleCloudPlatform/github-actions/tree/master/setup-gcloud) takes a service account email and key, pulls down `gcloud` and gets it configured, and installs credentials that are used by subsequent steps:

``` yaml
{{HTMLSafePassThrough `
- name: "GCP: setup-gcloud"
  uses: GoogleCloudPlatform/github-actions/setup-gcloud@master
  with:
    export_default_credentials: true
    project_id: passages-signup
    version: '285.0.0'

    service_account_email: ${{ secrets.GCPEmail }}
    service_account_key: ${{ secrets.GCPKey }}
`}}
```

Getting that email/key is the hardest part of the whole process. It involves jumping over to Google's web console, generating a new set of credentials for a "service account" [1], downloading them as a _JSON file_ (!!), passing its contents through `base64` to encode them to a portable format, and saving the result as a GitHub Actions secret. It's not too bad once you've been through the process a couple times, but is unexpectedly heavy compared to the simple API keys provided by most developer services.

But from there, subsequent Google Cloud commands are just invocations of `gloud`. This is _hugely_ helpful compared to some GitHub Actions modules because it means you can test any of them locally instead of deferring to slow/opaque CI loops. Here I send my project and its `Dockerfile` up to Google's cloud to be baked into a container image:

``` yaml
- name: "GCP: Publish image"
  run: gcloud builds submit --tag gcr.io/passages-signup/passages-signup
```

It's the exact command I used to run manually whenever I wanted to deploy. It now happens automatically from GitHub Actions.

These two take that image and deploy it to two separate Cloud Run apps (each of my newsletters has an independent deployment):

``` yaml
- name: "GCP: Deploy nanoglyph-signup"
  run: gcloud run deploy --image gcr.io/passages-signup/passages-signup
    --platform managed --region us-central1 nanoglyph-signup

- name: "GCP: Deploy passages-signup"
  run: gcloud run deploy --image gcr.io/passages-signup/passages-signup
    --platform managed --region us-central1 passages-signup
```

It was all pleasantly fast and easy to get working. I had my manual deploy recipe converted to an automated process in less than 10 minutes including troubleshooting, which is speed that's practically unheard of when it comes to plugging exotic new things into CI [2].

## The dependency graph (#dependency-graph)

Another minor nicety is GitHub Actions' easy system for specifying dependencies between jobs, which let me easily specify that the deployment job should only run if the `build` job finished successfully (note the use of the `needs` keyword):

``` yaml
jobs:
  build:
    steps:
      ...

  deploy-google-cloud-run:
    if: github.ref == 'refs/heads/master'
    needs: build
    steps:
      ...
```

Like I said, it's minor, but ensures that deployment only happens on a valid build, and furthermore cleanly encapsulates both the run output and configuration for each step into its own section.

[1] Synthetic jargon for a type of account that represents a process of some kind rather than a person.

[2] Getting new things up and running in CI tends to be especially slow because the development loop is: `git push`, wait 5 minutes, `git push`, wait 5 minutes, `git push`, etc. -- just about as slow as it gets.
