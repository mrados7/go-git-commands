# go-git-commands

## commit

Uses branch name to detrmine the commit message.
If the branch name is `FIX/FE-1234/branch-name` then the commit message will be `[FIX] [FE-1234] commit-message`.

### Install

```bash
brew install mrados7/main/commit
# or	
go install github.com/mrados7/go-git-commands/commit@latest
```


### Usage

```bash
commit
```

## checkout
Creates a new branch from the current branch using naming strategy `<type>/<ticket-id>/<branch-name>`.
Supported types are `FEAT`, `FIX`, `IMPR`, `OPS`.


### Install

```bash
brew install mrados7/main/checkout
# or	
go install github.com/mrados7/go-git-commands/checkout@latest
```

### Usage

```bash
checkout
```
