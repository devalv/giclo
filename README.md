# giclo

GitHub Liked repos cloner

[![Go Report Card](https://goreportcard.com/badge/github.com/devalv/giclo)](https://goreportcard.com/report/github.com/devalv/giclo)
[![CodeQL](https://github.com/devalv/giclo/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/devalv/giclo/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/devalv/giclo/branch/main/graph/badge.svg)](https://codecov.io/gh/devalv/giclo)

## Installation

1. Make sure that proper version of **Go** installed and ENVs are set.

```bash
wget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz
rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz
# add to .zshrc
export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"
```

1. Run **make** command to install all dev-utils.

```bash
make setup
```

## Project layout

Directory names and meanings
<https://github.com/golang-standards/project-layout/blob/master/README_ru.md>