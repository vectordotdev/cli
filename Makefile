# Timber CLI Makefile
#
# The contents of this file MUST be compatible with GNU Make 3.81,
# so do not use features or conventions introduced in later releases
# (for example, the ::= assignement operator)
#
# The Makefile for github-release was used as a basis for this file.
# The specific version can be found here:
#   https://github.com/c4milo/github-release/blob/6d2edc2/Makefile

build_dir := $(CURDIR)/build
dist_dir := $(CURDIR)/dist
s3_prefix := packages.timber.io/cli

exec := timber
github_repo := timberio/cli
version = $(shell cat VERSION)

.DEFAULT_GOAL := dist

.PHONY: clean
clean: clean-build clean-dist

.PHONY: clean-build
clean-build:
	@echo "Removing build files"
	rm -rf $(build_dir)

.PHONY: clean-dist
clean-dist:
	@echo "Removing distribution files"
	rm -rf $(dist_dir)

.PHONY: build
build: clean-build
	@echo "Creating build directory"
	mkdir -p $(build_dir)
	@echo "Building targets"
	@CGO_ENABLED=0 gox -ldflags "-X main.version=$(version)" \
		-osarch="darwin/amd64" \
		-osarch="freebsd/amd64" \
		-osarch="linux/amd64" \
		-osarch="linux/arm" \
		-osarch="linux/arm64" \
		-osarch="netbsd/amd64" \
		-osarch="openbsd/amd64" \
		-output "$(build_dir)/$(exec)-$(version)-{{.OS}}-{{.Arch}}/$(exec)/bin/$(exec)"
	@for f in $$(ls $(build_dir)); do \
		readme_source="$(CURDIR)/README.md"; \
		readme_dest="$(build_dir)/$$f/$(exec)/"; \
		echo "Copying $$readme_source into $$readme_dest"; \
		cp $$readme_source $$readme_dest; \
		changelog_source="$(CURDIR)/CHANGELOG.md"; \
		changelog_dest="$(build_dir)/$$f/$(exec)/"; \
		echo "Copying $$changelog_source into $$changelog_dest"; \
		cp $$changelog_source $$changelog_dest; \
		license_source="$(CURDIR)/LICENSE"; \
		license_dest="$(build_dir)/$$f/$(exec)/"; \
		echo "Copying $$license_source into $$license_dest"; \
		cp $$license_source $$license_dest; \
	done

.PHONY: dist
dist: clean-dist build
	@echo "Creating distribution directory"
	mkdir -p $(dist_dir)
	@echo "Creating distribution archives"
	$(eval FILES := $(shell ls $(build_dir)))
	@for f in $(FILES); do \
		echo "Creating distribution archive for $$f"; \
		(cd $(build_dir)/$$f && tar -czf $(dist_dir)/$$f.tar.gz *); \
	done

.PHONY: release
release: dist
	@tag=v$(version); \
	commit=$(git rev-list -n 1 $$tag); \
	name=$$(git show -s $$tag --pretty=tformat:%N | sed -e '4q;d'); \
	changelog=$$(git show -s $$tag --pretty=tformat:%N | sed -e '1,5d'); \
	grease create-release --name "$$name" --notes "$$changelog" --assets "dist/*" $(github_repo) "$$tag" "$$commit"
	$(eval FILES := $(shell ls $(dist_dir)))
	@for exact_filename in $(FILES); do \
		rel=$$(echo $$exact_filename | sed "s/\.tar\.gz//"); \
		doublet=$$(echo $$rel | cut -d - -f 4,5); \
		latest_patch_version="$$(echo $(version) | cut -d . -f 1,2).x"; \
		latest_minor_version="$$(echo $(version) | cut -d . -f 1).x.x"; \
		latest_patch_filename=$$(echo $$exact_filename | sed "s/$(version)/$$latest_patch_version/"); \
		latest_minor_filename=$$(echo $$exact_filename | sed "s/$(version)/$$latest_minor_version/"); \
		exact_version_destination="s3://$(s3_prefix)/$(version)/$$doublet/$$exact_filename"; \
		latest_patch_destination="s3://$(s3_prefix)/$$latest_patch_version/$$doublet/$$latest_patch_filename"; \
		latest_minor_destination="s3://$(s3_prefix)/$$latest_minor_version/$$doublet/$$latest_minor_filename"; \
		echo "Uploading v$(version) as $(version) for $$arch ($$exact_filename) to S3 ($$exact_version_destination)"; \
		aws s3 cp $(dist_dir)/$$exact_filename $$exact_version_destination; \
		echo "Uploading v$(version) as $$latest_patch_version for $$arch ($$exact_filename) to S3 ($$latest_patch_destination)"; \
		aws s3 cp $(dist_dir)/$$exact_filename $$latest_patch_destination; \
		echo "Uploading v$(version) as $$latest_minor_version for $$arch ($$exact_filename) to S3 ($$latest_patch_destination)"; \
		aws s3 cp $(dist_dir)/$$exact_filename $$latest_minor_destination; \
	done

.PHONY: get-tools
get-tools:
	go get github.com/golang/dep/cmd/dep
	go get github.com/mitchellh/gox
	go get github.com/jstemmer/go-junit-report

.PHONY: test
test:
	@go test -v