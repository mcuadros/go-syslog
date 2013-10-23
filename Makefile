help:
	@echo "Available targets:"
	@echo "- test: run tests"
	@echo "- installdependencies: installs dependencies declared in dependencies.txt"
	@echo "- clean: cleans directory"

installdependencies:
	cat dependencies.txt | xargs go get

test: installdependencies
	go test -i && go test

clean:
	find . -type 'f' -name '*.test' -print | xargs rm -f
