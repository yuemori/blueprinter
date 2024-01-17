.PHONY: build example-app1

build:
	go build -v -o blueprinter

example-app1: build
	./blueprinter generate --ignore=example/app1/*.go --out=./example/app1/container/container.generated.go --workdir=example/app1 github.com/yuemori/blueprinter/example/app1/container Container
