#!/bin/sh

kubectx=/usr/bin/kubectx
go_kubectx="$GOPATH/bin/kubectx"

kubens=/usr/bin/kubens
go_kubens="$GOPATH/bin/kubens"

bench_kubectx() {
  echo "COMMAND: kubectx $@"
  hyperfine "$kubectx $@" "$go_kubectx $@" --export-markdown bench_kubectx.md 2>/dev/null
  printf "\n\n"
}

bench_kubens() {
  echo "COMMAND: kubens $@"
  hyperfine "$kubens $@" "$go_kubens $@" --export-markdown bench_kubens.md 2>/dev/null
  printf "\n\n"
}

bench_kubectx minikube
bench_kubens kube-system

