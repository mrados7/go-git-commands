# go-git-commands

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

## Config file

You can add a config file to project directory `.git-commands.json` to override the default values.
Alternatively, you can add a global config file to your home directory `~/.git-commands.json`.

```json
{
  "branchTypes": [
    {
      "type": "FEAT",
      "description": "A new feature"
    },
    {
      "type": "FIX",
      "description": "A bug fix"
    },
    {
      "type": "DOCS",
      "description": "Documentation only changes"
    },
    ...
  ],
  "boards": [
    {
      "name": "PTB",
      "description": "Platform Team Board"
    },
    {
      "name": "SDB",
      "description": "Support Desk Board"
    }
    ...
  ]
}

```
