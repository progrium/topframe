
build:
	go build -o local/topframe .

release:
	goreleaser --snapshot --skip-publish --rm-dist

clean:
	rm -rf ./dist