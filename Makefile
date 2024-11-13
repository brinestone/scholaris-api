dev:
	@encore run

test:
	@encore test ./...

test_watch:
	@$$GOPATH/bin/air --build.bin "encore -v test ./..." --build.exclude_dir ".encore,node_modules"
