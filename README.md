# sorg

A statically compiled version of [org][org].

## Build

Install Go 1.6, set up and run [black-swan][black-swan], then:

    cp .env.sample
    make install
    make build

The project can be deployed to s3 using:

    AWS_ACCESS_KEY_ID=...
    AWS_SECRET_ACCESS_KEY=...
    S3_BUCKET=...
    make deploy

[black-swan]: https://github.com/brandur/black-swan
[org]: https://github.com/brandur/org
