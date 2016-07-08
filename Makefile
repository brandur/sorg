all: clean install test vet lint check-gofmt build

build:
	$(GOPATH)/bin/sorg-build

check-gofmt:
	scripts/check_gofmt.sh

clean:
	mkdir -p public/
	rm -f -r public/*

deploy: build
# Note that AWS_ACCESS_KEY_ID will only be set for builds on the master
# branch because it's stored in `.travis.yml` as an encrypted variable.
# Encrypted variables are not made available to non-master branches because
# of the risk of being leaked through a script in a rogue pull request.
ifdef AWS_ACCESS_KEY_ID
	aws --version

	# Force text/html for HTML because we're not using an extension.
	#
	# Note that we don't delete because it could result in a race condition in
	# that files that are uploaded with special directives below could be
	# removed even while the S3 bucket is actively in-use.
	aws s3 sync ./public/ s3://$(S3_BUCKET)/ --acl public-read --content-type text/html --exclude 'assets*' $(AWS_CLI_FLAGS)

	# Then move on to assets and allow S3 to detect content type.
	aws s3 sync ./public/assets/ s3://$(S3_BUCKET)/assets/ --acl public-read --delete --follow-symlinks $(AWS_CLI_FLAGS)

	# This one is a bit tricker to explain, but what we're doing here is
	# uploading directory indexes as files at their directory name. So for
	# example, 'articles/index.html` gets uploaded as `articles`.
	#
	# We do all this work because CloudFront/S3 has trouble with index files.
	# An S3 static site can have index.html set to indexes, but CloudFront only
	# has the notion of a "root object" which is an index at the top level.
	#
	# We do this during deploy instead of during build for two reasons:
	#
	# 1. Some directories need to have an index *and* other files. We must name
	#    index files with `index.html` locally though because a file and
	#    directory cannot share a name.
	# 2. The `index.html` files are useful for emulating a live server locally:
	#    Golang's http.FileServer will respect them as indexes.
	find ./public -name index.html | egrep -v './public/index.html' | sed "s|^\./public/||" | xargs -I {} -n 1 dirname {} | xargs -I {} -n 1 aws s3 cp ./public/{}/index.html s3://$(S3_BUCKET)/{} --acl public-read --content-type text/html
else
	# No AWS access key. Skipping deploy.
endif

install:
	go install $(shell go list ./... | egrep -v '/org/|/vendor/')

lint:
	$(GOPATH)/bin/golint

	# Hack to workaround the fact that Golint doesn't produce a non-zero exit
	# code on failure because Go Core team is always right and everyone else is
	# always wrong:
	#
	#     https://github.com/golang/lint/issues/65
	#
	test -z "$$(golint .)"

save:
	godep save $(shell go list ./... | egrep -v '/org/|/vendor/')

serve:
	$(GOPATH)/bin/sorg-serve

test:
	go test $(shell go list ./... | egrep -v '/org/|/vendor/')

vet:
	go vet $(shell go list ./... | egrep -v '/org/|/vendor/')

watch:
	fswatch -o layouts/ org/ views/ | xargs -n1 -I{} make build
