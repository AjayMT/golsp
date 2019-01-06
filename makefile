
golsp: *.go core/*.go stdlib/**/*
	make -C stdlib/types
	make -C stdlib/os
	make -C stdlib/stream
	go build -o golsp *.go

.PHONY: clean
clean:
	rm -f golsp
	rm -f stdlib/**/*.so
