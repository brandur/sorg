all: clean install test vet lint check-gofmt build

build:
	$(GOPATH)/bin/sorg-build

check-gofmt:
	scripts/check_gofmt.sh

clean:
	mkdir -p public/
	rm -f -r public/*

# Requires that variables necessary to make an AWS API call are in the
# environment.
check-aws-keys:
ifndef AWS_ACCESS_KEY_ID
	$(error AWS_ACCESS_KEY_ID is required)
endif
ifndef AWS_SECRET_ACCESS_KEY
	$(error AWS_SECRET_ACCESS_KEY is required)
endif

# Requires that variables necessary to update a CloudFront distribution are in
# the environment.
check-cloudfront-id:
ifndef CLOUDFRONT_ID
	$(error CLOUDFRONT_ID is required)
endif

# Long TTL (in seconds) to set on an object in S3. This is suitable for items
# that we expect to only have to invalidate very rarely like images. Although
# we set it for all assets, those that are expected to change more frequently
# like script or stylesheet files are versioned by a path that can be set at
# build time.
LONG_TTL := 86400

# Short TTL (in seconds) to set on an object in S3. This is suitable for items
# that are expected to change more frequently like any HTML file.
SHORT_TTL := 3600

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
	aws s3 sync ./public/ s3://$(S3_BUCKET)/ --acl public-read --cache-control max-age=$(SHORT_TTL) --content-type text/html --exclude 'assets*' $(AWS_CLI_FLAGS)

	# Then move on to assets and allow S3 to detect content type.
	aws s3 sync ./public/assets/ s3://$(S3_BUCKET)/assets/ --acl public-read --cache-control max-age=$(LONG_TTL) --delete --follow-symlinks $(AWS_CLI_FLAGS)

	# Upload Atom feed files with their proper content type.
	find ./public -name '*.atom' | sed "s|^\./public/||" | xargs -I{} -n1 aws s3 cp ./public/{} s3://$(S3_BUCKET)/{} --acl public-read --cache-control max-age=$(SHORT_TTL) --content-type application/xml

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
	find ./public -name index.html | egrep -v './public/index.html' | sed "s|^\./public/||" | xargs -I{} -n1 dirname {} | xargs -I{} -n1 aws s3 cp ./public/{}/index.html s3://$(S3_BUCKET)/{} --acl public-read --cache-control max-age=$(SHORT_TTL) --content-type text/html
else
	# No AWS access key. Skipping deploy.
endif

install:
	go install $(shell go list ./... | egrep -v '/vendor/')

# Invalidates CloudFront's cache for paths specified in PATHS.
#
# Usage:
#     make PATHS="/fragments /fragments/six-weeks" invalidate
invalidate: check-aws-keys check-cloudfront-id
ifndef PATHS
	$(error PATHS is required)
endif
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths ${PATHS}

# Invalidates CloudFront's entire cache.
invalidate-all: check-aws-keys check-cloudfront-id
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths /

# Invalidates CloudFront's cached assets.
invalidate-assets: check-aws-keys check-cloudfront-id
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths /assets

# Invalidates CloudFront's cached index pages. This is useful, but not
# necessarily required, when publishing articles or new data (if it's not run,
# anything cached in CloudFront will expire naturally after SHORT_TTL).
invalidate-indexes: check-aws-keys check-cloudfront-id
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths /articles /articles.atom /fragments /fragments.atom /photos /reading /runs /twitter

# Note that unfortunately Golint doesn't work like other Go commands: it only
# takes only a single argument at a time and expects that each is the name of a
# local directory (as opposed to a package).
#
# The exit 255 trick ensures that xargs will actually bubble a failure back up
# to the entire command.
lint:
	go list ./... | egrep -v '/vendor/' | sed "s|^github\.com/brandur/sorg|.|" | xargs -I{} -n1 sh -c '$(GOPATH)/bin/golint -set_exit_status {} || exit 255'

save:
	godep save $(shell go list ./... | egrep -v '/vendor/')

serve:
	$(GOPATH)/bin/sorg-serve

test:
	psql postgres://localhost/sorg-test < testing/black_swan.sql > /dev/null
	go test $(shell go list ./... | egrep -v '/vendor/')

vet:
	go vet $(shell go list ./... | egrep -v '/vendor/')

# Note that we use the CONTENT_ONLY flag on the build here. We're watching for
# changes in content directories, so don't bother rebuilding pages generated
# from Black Swan data.
watch:
	fswatch -o content/ layouts/ pages/ views/ | CONTENT_ONLY=true xargs -n1 -I{} make build

# This is designed to be compromise between being explicit and readability. We
# can allow the find to discover everything in vendor/, but then the fswatch
# invocation becoems a huge unreadable wall of text that gets dumped into the
# shell. Instead, find all our own *.go files and then just tack the vendor/
# directory on separately (fswatch will watch it recursively).
GO_FILES := $(shell find . -type f -name "*.go" ! -path "./org/*" ! -path "./vendor/*")

# We recompile our Go source when a file changes, but we also rebuild the site
# because a change in source may have affected the build formula.
watch-go:
	fswatch -o $(GO_FILES) vendor/ | xargs -n1 -I{} make install build
