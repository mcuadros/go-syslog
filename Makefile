help:
	@echo "Available targets:"
	@echo "- test: run tests"
	@echo "- installdependencies: installs dependencies declared in dependencies.txt"
	@echo "- clean: cleans directory"
	@echo "- benchmarks: run benchmarks"

installdependencies:
	cat dependencies.txt | xargs go get

test: installdependencies
	go test -i && go test

clean:
	find . -type 'f' -name '*.test' -print | xargs rm -f

benchmarks:
	go test -gocheck.b
