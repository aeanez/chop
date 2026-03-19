# Shell Completions

chop provides tab-completion scripts for bash, zsh, fish, and PowerShell. Completions cover all top-level commands, subcommands, and flags.

## bash

Add to `~/.bashrc`:

```bash
source <(chop completion bash)
```

Or install permanently (no subprocess on each shell start):

```bash
chop completion bash > ~/.local/share/bash-completion/completions/chop
```

## zsh

Add to `~/.zshrc`:

```zsh
source <(chop completion zsh)
```

Or install to your completions directory:

```zsh
chop completion zsh > "${fpath[1]}/_chop"
```

## fish

Add to your fish config (`~/.config/fish/config.fish`):

```fish
chop completion fish | source
```

Or install permanently:

```fish
chop completion fish > ~/.config/fish/completions/chop.fish
```

## PowerShell

Add to your `$PROFILE`:

```powershell
chop completion powershell | Invoke-Expression
```

To find your profile path: `echo $PROFILE`
