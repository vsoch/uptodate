# Getting Started

Getting started with development is easy! First clone the repository, optionally into 
your src/github.com/vsoch `$GOPATH`:

```bash
$ git clone https://github.com/vsoch/uptodate
$ cd uptodate
```

And then you can easily use the Makefile to also just build or run:

```bash
$ make

# This won't include formatting to change the files
$ make build
```

or you can use go directly!

```bash
$ go run main.go dockerfile 
```
