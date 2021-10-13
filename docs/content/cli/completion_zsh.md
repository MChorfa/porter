---
title: "porter completion zsh"
slug: porter_completion_zsh
url: /cli/porter_completion_zsh/
---
## porter completion zsh

generate the autocompletion script for zsh

### Synopsis


Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:
# Linux:
$ porter completion zsh > "${fpath[1]}/_porter"
# macOS:
$ porter completion zsh > /usr/local/share/zsh/site-functions/_porter

You will need to start a new shell for this setup to take effect.


```
porter completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
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

