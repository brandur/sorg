+++
hook = "A workaround to get nice stack traces for errors in Go."
published_at = 2021-08-22T21:56:16Z
title = "Error stack traces in Go with x/xerrors"
+++

Go's last standing major weakness is error handling. A few years ago the list was much longer, with the language missing an adequate package manager, system for pulling static assets into a binary, and generics. But now, the first two have already been addressed with [Go Modules](https://go.dev/blog/using-go-modules) in 1.11 and [`go:embed` in 1.16](https://pkg.go.dev/embed), and generics are expected to be in beta form by Go 1.18's release in December. Errors are the last major omission.

Go 1.13 brought in some improvements in the form of the fmt-er `%w` symbol and `errors.Is`, `.As`, and `.Unwrap`, but still outstanding are (1) a way to reasonably get a stack trace at the site where an error is generated, and (2) a way to cut down on `if err != nil { ... }` boilerplate that pervades all Go code.

Our API's set up so that if an error is returned back to common API endpoint infrastructure by a specific handler, or if a panic occurs, we push an event to Sentry, where we'll be emailed about it, and which will act as a convenient triage bucket of potential problems. This works well, but we've been a little frustrated by how context-free the errors there are. We get a stack trace, but it's the stack trace of the line of code where we emit to Sentry, making it only noise. _Hopefully_ someone wrapped an error with a greppable string, because otherwise figuring out where an error was generated involves a lot of guesswork.

This doesn't seem like something that'll be addressed in the core language anytime soon, so we implemented a workaround. The [`x/xerrors` package](https://pkg.go.dev/golang.org/x/xerrors) was a nursery for the error additions that eventually landed in 1.13. But unlike 1.13's errors, `xerrors` also includes a feature that was _not_ brought in -- stack traces. An xerror will capture the call site where it was generated, and the Go packages for services like Sentry are smart enough to recognize an xerror and push up its metadata.

## xerrors migration (#migration)

Luckily, migration was very easy, We converted every instance of `fmt.Errorf` over to `xerrors.Errorf` in just two commands using Gofmt's rewrite parameter along with an invocation of [Goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports):

``` sh
gofmt -w -r 'fmt.Errorf -> xerrors.Errorf' .
goimports -w .
```

Gofmt rewrites the AST and Goimports corrects import paths by removing `fmt` and adding `xerrors` as appropriate. One neat thing about the use of Gofmt here instead of find + replace is that because it's looking at the AST, Gofmt will pick up `fmt.Errorf` calls that were strangely formatted, e.g. broken up across multiple lines.

We then forbid future uses of `fmt.Errorf` with [golangci-lint](https://github.com/golangci/golangci-lint) and forbidigo:

``` yaml
linters:
  enable:
    - forbidigo

linters-settings:
  forbidigo:
    forbid:
      - '^fmt\.Errorf$'%
```

## Error handling convention (#convention)

Our internal convention for how to do error handling is now:

* At boundaries between our code and calls out to external packages, make an effort to always wrap the result with `xerrors.Errorf`. This ensures that we always capture a stack trace at the most proximate location of an error being generated as possible. For example, I'd always wrap an invocation out to sqlc:

    ``` go
    _, err = queries.ClusterInsert(ctx, dbsqlc.ClusterInsertParams{
        ...
    })
    if err != nil {
        return nil, xerrors.Errorf("error inserting cluster: %w", err)
    }
    ```

* When handling errors being returned from our own code, it's okay to wrap errors further with `xerrors.Errorf` if the extra context is useful, but not imperative.

This end solution isn't _amazing_, but it works. We'd of course far prefer if Go itself had a built-in equivalent. Reaching out to ex-colleagues at other companies, although there are roughly equivalent alternatives to what we're doing with `xerrors`, no one seems to have anything all that much better.
