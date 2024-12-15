# Go tool


```sh
# install release candidate
go install golang.org/dl/go1.24rc1@latest
go1.24rc download

# practical alias
alias go=go1.24rc

# update go version in go.mod
go get go@1.24rc1
# add tools
go get -tool mvdan.cc/gofumpt
go get -tool github.com/google/osv-scanner/cmd/osv-scanner

# list tools
go tool

# update tools
go get -u tool
```


## Docs

* https://tip.golang.org/doc/go1.24#tools)
* https://tip.golang.org/doc/modules/managing-dependencies#tools
