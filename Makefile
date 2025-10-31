.PHONY: release-patch
release-patch:
	git add .
	$(eval CURRENT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"))
	$(eval NEW_VERSION=$(shell echo $(CURRENT_VERSION) | awk -F. -v OFS=. '{$$NF++;} 1'))
	git commit --no-verify --allow-empty -m "patch release $(NEW_VERSION)"
	git push origin main
	git tag $(NEW_VERSION)
	git push origin $(NEW_VERSION)
	go install github.com/commoddity/devint@$(NEW_VERSION)

.PHONY: release-minor
release-minor:
	git add .
	$(eval CURRENT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"))
	$(eval NEW_VERSION=$(shell echo $(CURRENT_VERSION) | awk -F. -v OFS=. '{$$2++; $$NF=0} 1'))
	git commit --no-verify --allow-empty -m "minor release $(NEW_VERSION)"
	git push origin main
	git tag $(NEW_VERSION)
	git push origin $(NEW_VERSION)
	go install github.com/commoddity/devint@$(NEW_VERSION)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/main.exe main.go
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/main main.go
build-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/main main.go
