dev:
	@encore run

test:
	@encore -v test ./...

test_watch:
	@$$GOPATH/bin/air --build.bin "make test" --build.exclude_dir ".encore"