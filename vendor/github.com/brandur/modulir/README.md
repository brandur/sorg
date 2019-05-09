# Modulir [![Build Status](https://travis-ci.org/brandur/modulir.svg?branch=master)](https://travis-ci.org/brandur/modulir)

Modulir is an experimental mini-framework for static site
generation that suggests that a site's main build recipe
should be written in Go, both for type safety and to
provide as much flexibility as possible.

The main features that the package provides are an entry
point that takes a build loop, a job pool for enqueuing the
set of parallel jobs that make up the build, and a set of
modules that provide helpers for various features in Go's
core and in other useful libraries that would otherwise be
quite verbose.

The package is currently highly experimental and its APIs
subject to change.

<!--
# vim: set tw=79:
-->
