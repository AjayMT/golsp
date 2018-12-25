
golsp: *.go core/*.go stdlib/**/*
	make -C stdlib/types
	go build -o golsp *.go

.PHONY: clean
clean:
	rm -f golsp
	rm -f stdlib/**/*.so
