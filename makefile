
golsp: src/**/*.go src/*.go
	go build -o golsp src/cli.go

.PHONY: clean
clean:
	rm -f golsp
