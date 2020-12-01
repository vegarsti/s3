# s3

<p>
    <a href="https://github.com/vegarsti/s3/releases"><img src="https://img.shields.io/github/release/vegarsti/s3.svg" alt="Latest Release"></a>
    <a href="https://github.com/vegarsti/s3/actions"><img src="https://github.com/vegarsti/s3/workflows/build/badge.svg" alt="Build Status"></a>
    <a href="http://goreportcard.com/report/github.com/vegarsti/s3"><img src="http://goreportcard.com/badge/vegarsti/s3" alt="Go ReportCard"></a>
</p>

## Installation

```sh
$ go get "github.com/vegarsti/s3"
```

## Usage

Set environment variables `AWS_REGION` and `AWS_BUCKET`.

```sh
$ echo "hello" > hello.txt

$ s3 upload hello.txt

$ s3 list
hi.txt
hello.txt

$ rm hello.txt

$ s3 download hello.txt

$ cat hello.txt
hello

$ s3 delete hello.txt

$ s3 list
hi.txt

```
