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
	aws s3 sync ./public/ s3://$(S3_BUCKET)/ --acl public-read --content-type text/html --delete --exclude 'assets*' $(AWS_CLI_FLAGS)

	# Then move on to assets and allow S3 to detect content type.
	aws s3 sync ./public/assets/ s3://$(S3_BUCKET)/assets/ --acl public-read --delete --follow-symlinks $(AWS_CLI_FLAGS)
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
	fswatch -o org/ | xargs -n1 -I{} make build
