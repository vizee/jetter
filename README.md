# jetter

cli for [jet](https://github.com/CloudyKit/jet)

syntax: https://github.com/CloudyKit/jet/blob/master/docs/syntax.md

## Installation

```
go install github.com/vizee/jetter@latest
```

## Usage

run with files:

```
cd testdata/templates/hello
jetter --set name=world --set name2=WORLD . hello.jet hello2.jet
```

run with a directory:

```
jetter -d testdata/templates -v testdata/example-values.yaml -o testdata/output -e .txt example
```
