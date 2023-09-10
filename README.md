# giclo

GitHub Liked repos cloner

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
