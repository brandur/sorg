.PHONY: all
all: clean install test vet lint check-dl0 check-gofmt check-headers check-retina build

.PHONY: build
build:
	$(shell go env GOPATH)/bin/sorg build

.PHONY: check-d10
check-dl0:
	scripts/check_dl0.sh

.PHONY: check-gofmt
check-gofmt:
	scripts/check_gofmt.sh

.PHONY: check-headers
check-headers:
	scripts/check_headers.sh

.PHONY: check-retina
check-retina:
	scripts/check_retina.sh

.PHONY: clean
clean:
	mkdir -p public/
	rm -f -r public/*

.PHONY: compile
compile: install

# Updates TOML flat files in data/ from qself, which is running on regular cron
# to fetch updates from these services.
.PHONY: data-update
data-update:
	curl --compressed --output data/goodreads.toml https://raw.githubusercontent.com/brandur/qself-brandur/master/data/goodreads.toml
	curl --compressed --output data/twitter.toml https://raw.githubusercontent.com/brandur/qself-brandur/master/data/twitter.toml

# data-update aliases
.PHONY: update-data
update-data: data-update

# Long TTL (in seconds) to set on an object in S3. This is suitable for items
# that we expect to only have to invalidate very rarely like images. Although
# we set it for all assets, those that are expected to change more frequently
# like script or stylesheet files are versioned by a path that can be set at
# build time.
LONG_TTL := 86400

# Short TTL (in seconds) to set on an object in S3. This is suitable for items
# that are expected to change more frequently like any HTML file.
SHORT_TTL := 3600

.PHONY: deploy
deploy: check-target-dir
# Note that AWS_ACCESS_KEY_ID will only be set for builds on the master branch
# because it's stored in GitHub as a secret variable. Secret variables are not
# made available to non-master branches because of the risk of being leaked
# through a script in a rogue pull request.
ifdef AWS_ACCESS_KEY_ID
	aws --version

	@echo "\n=== Syncing HTML files\n"

	# Force text/html for HTML because we're not using an extension.
	#
	# Note that we don't delete because it could result in a race condition in
	# that files that are uploaded with special directives below could be
	# removed even while the S3 bucket is actively in-use.
	aws s3 sync $(TARGET_DIR) s3://$(S3_BUCKET)/ --acl public-read --cache-control max-age=$(SHORT_TTL) --content-type text/html --exclude 'assets*' --exclude 'photographs*' $(AWS_CLI_FLAGS)

	@echo "\n=== Syncing media assets\n"

	# Then move on to assets and allow S3 to detect content type.
	#
	# Note use of `--size-only` because mtimes may vary as they're not
	# preserved by Git. Any updates to a static asset are likely to change its
	# size though.
	aws s3 sync $(TARGET_DIR)/assets/ s3://$(S3_BUCKET)/assets/ --acl public-read --cache-control max-age=$(LONG_TTL) --follow-symlinks --size-only $(AWS_CLI_FLAGS)

	@echo "\n=== Syncing photographs\n"

	# Photographs are identical to assets above except without `--delete`
	# because any given build probably doesn't have the entire set.
	aws s3 sync $(TARGET_DIR)/photographs/ s3://$(S3_BUCKET)/photographs/ --acl public-read --cache-control max-age=$(LONG_TTL) --follow-symlinks --size-only $(AWS_CLI_FLAGS)

	@echo "\n=== Syncing Atom feeds\n"

	# Upload Atom feed files with their proper content type.
	find $(TARGET_DIR) -name '*.atom' | sed "s|^\$(TARGET_DIR)/||" | xargs -I{} -n1 aws s3 cp $(TARGET_DIR)/{} s3://$(S3_BUCKET)/{} --acl public-read --cache-control max-age=$(SHORT_TTL) --content-type application/xml

	@echo "\n=== Syncing index HTML files\n"

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
	find $(TARGET_DIR) -name index.html | egrep -v '$(TARGET_DIR)/index.html' | sed "s|^$(TARGET_DIR)/||" | xargs -I{} -n1 dirname {} | xargs -I{} -n1 aws s3 cp $(TARGET_DIR)/{}/index.html s3://$(S3_BUCKET)/{} --acl public-read --cache-control max-age=$(SHORT_TTL) --content-type text/html

	@echo "\n=== Fixing robots.txt content type\n"

	# Give robots.txt (if it exists) a Content-Type of text/plain. Twitter is
	# rabid about this.
	[ -f $(TARGET_DIR)/robots.txt ] && aws s3 cp $(TARGET_DIR)/robots.txt s3://$(S3_BUCKET)/ --acl public-read --cache-control max-age=$(SHORT_TTL) --content-type text/plain $(AWS_CLI_FLAGS) || echo "no robots.txt"

	@echo "\n=== Setting redirects\n"

	# Set redirects on specific objects. Generally old stuff that I ended up
	# refactoring to live somewhere else. Note that for these to work, S3 web
	# hosting must be on, and CloudFront must be pointed to the S3 web hosting
	# URL rather than the REST endpoint.
	aws s3api put-object --acl public-read --bucket $(S3_BUCKET) --key glass --website-redirect-location /newsletter

else
	# No AWS access key. Skipping deploy.
endif

.PHONY: install
install:
	go install .

# Invalidates CloudFront's cache for paths specified in PATHS.

# Usage:
#     make PATHS="/fragments /fragments/six-weeks" invalidate
.PHONY: invalidate
invalidate: check-aws-keys check-cloudfront-id
ifndef PATHS
	$(error PATHS is required)
endif
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths ${PATHS}

# Invalidates CloudFront's entire cache.
.PHONY: invalidate-all
invalidate-all: check-aws-keys check-cloudfront-id
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths /

# Invalidates CloudFront's cached assets.
.PHONY: invalidate-assets
invalidate-assets: check-aws-keys check-cloudfront-id
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths /assets

# Invalidates CloudFront's cached index pages. This is useful, but not
# necessarily required, when publishing articles or new data (if it's not run,
# anything cached in CloudFront will expire naturally after SHORT_TTL).
.PHONY: invalidate-indexes
invalidate-indexes: check-aws-keys check-cloudfront-id
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths /articles /articles.atom /fragments /fragments.atom /nanoglyphs /nanoglyphs.atom /now /passages /passages.atom /twitter

.PHONY: killall
killall:
	killall sorg

.PHONY: lint
lint:
	$(shell go env GOPATH)/bin/golint -set_exit_status ./...

.PHONY: loop
loop:
	$(shell go env GOPATH)/bin/sorg loop

# A specialized S3 bucket used only for caching resized photographs.
PHOTOGRAPHS_S3_BUCKET := "brandur.org-photographs"

.PHONY: photographs-download
photographs-download:
ifdef AWS_ACCESS_KEY_ID
	aws s3 sync s3://$(PHOTOGRAPHS_S3_BUCKET)/ content/photographs/
else
	# No AWS access key. Skipping photographs-download.
endif

.PHONY: photographs-download-markers
photographs-download-markers:
ifdef AWS_ACCESS_KEY_ID
	aws s3 sync s3://$(PHOTOGRAPHS_S3_BUCKET)/ content/photographs/ --exclude "*" --include "*.marker"
else
	# No AWS access key. Skipping photographs-download-markers.
endif

.PHONY: photographs-upload
photographs-upload:
ifdef AWS_ACCESS_KEY_ID
	aws s3 sync content/photographs/ s3://$(PHOTOGRAPHS_S3_BUCKET)/ --size-only
else
	# No AWS access key. Skipping photographs-upload.
endif

.PHONY: sigusr2
sigusr2:
	killall -SIGUSR2 sorg

# sigusr2 aliases
.PHONY: reboot
reboot: sigusr2
.PHONY: restart
restart: sigusr2

.PHONY: test
test:
	go test ./...

.PHONY: test-nocache
test-nocache:
	go test -count=1 ./...

.PHONY: vet
vet:
	go vet ./...

# This is designed to be compromise between being explicit and readability. We
# can allow the find to discover everything in vendor/, but then the fswatch
# invocation becomes a huge unreadable wall of text that gets dumped into the
# shell. Instead, find all our own *.go files and then just tack the vendor/
# directory on separately (fswatch will watch it recursively).
GO_FILES := $(shell find . -type f -name "*.go" ! -path "./vendor/*")

# Meant to be used in conjuction with `forego start`. When a Go file changes,
# this watch recompiles the project, then sends USR2 to the process which
# prompts Modulir to re-exec it.
.PHONY: watch-go
watch-go:
	fswatch -o $(GO_FILES) vendor/ | xargs -n1 -I{} make install sigusr2

#
# Helpers
#

# Requires that variables necessary to make an AWS API call are in the
# environment.
.PHONY: check-aws-keys
check-aws-keys:
ifndef AWS_ACCESS_KEY_ID
	$(error AWS_ACCESS_KEY_ID is required)
endif
ifndef AWS_SECRET_ACCESS_KEY
	$(error AWS_SECRET_ACCESS_KEY is required)
endif

# Requires that variables necessary to update a CloudFront distribution are in
# the environment.
.PHONY: check-cloudfront-id
check-cloudfront-id:
ifndef CLOUDFRONT_ID
	$(error CLOUDFRONT_ID is required)
endif

.PHONY: check-target-dir
check-target-dir:
ifndef TARGET_DIR
	$(error TARGET_DIR is required)
endif
