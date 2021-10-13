---
title: "porter completion fish"
slug: porter_completion_fish
url: /cli/porter_completion_fish/
---
## porter completion fish

generate the autocompletion script for fish

### Synopsis


Generate the autocompletion script for the fish shell.

To load completions in your current shell session:
$ porter completion fish | source

To load completions for every new session, execute once:
$ porter completion fish > ~/.config/fish/completions/porter.fish

You will need to start a new shell for this setup to take effect.


```
porter completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter completion](/cli/porter_completion/)	 - generate the autocompletion script for the specified shell

