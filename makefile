
golsp: *.go
	go build -o golsp *.go

.PHONY: clean
clean:
	rm -f golsp
