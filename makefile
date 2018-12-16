
golsp: src/**/*.go src/*.go
	go build -o golsp src/*.go

.PHONY: clean
clean:
	rm -f golsp
