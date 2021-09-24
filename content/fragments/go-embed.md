+++
hook = "A handy trick to get all the advantages of self-contained binaries with `go:embed`, while keeping development loops fast in development."
published_at = 2021-09-24T16:59:41Z
title = "go:embed in prod, serve-from-disk in development"
+++

One of Go 1.16's best new features was [`go:embed`](https://golang.org/doc/go1.16), which allows the compile process to easily slurp up files from the local file system into built binaries.

It's useful in many ways, but one of my favorites is in small Go web servers that use templates, and need to serve static assets like CSS and images. Previously, Go alone was insufficient for getting something like that deployed -- you'd need to layer a container on top which had the requisite files copied in.

`go:embed` solves that definitively, providing an easy directive for pulling in files, and making itself interoperable with other Go built-ins:

``` go
//go:embed static/*.css static/*.js static/*.png
var staticAssets embed.FS

http.Handle("/static/",
    http.StripPrefix("/static/", http.FileServer(http.FS(staticAssets))))
```

And voila -- you can ship that binary anywhere. No containers required.

This works beyond HTML servers too. The main product I work on is an API that doesn't have a whole lot of web-flavored static assets, but it depends on a few YAML files containing configuration and seed data. After 1.16 we put those into `go:embed`, and it's now a binary with zero dependencies.

## Dev/prod asymmetry (#asymmetry)

Another trick is to use `go:embed` for production deployment, but assets from the local filesystem otherwise:

``` go
//go:embed static/*.css static/*.js static/*.png
var staticAssets embed.FS

var fileSystem http.FileSystem
if isProduction {
    fileSystem = http.FS(staticAssets)
} else {
    fileSystem = http.Dir("./static")
}

http.Handle("/static/",
    http.StripPrefix("/static/", http.FileServer(fileSystem)))
```

This setup enables faster iteration in development when you're still changing a lot of things. Instead of rebuilding every time you make a change, you start the binary once, change CSS/JS/templates as required, and refresh to see the results immediately.
