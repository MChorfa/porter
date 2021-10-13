---
title: "porter completion bash"
slug: porter_completion_bash
url: /cli/porter_completion_bash/
---
## porter completion bash

generate the autocompletion script for bash

### Synopsis


Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:
$ source <(porter completion bash)

To load completions for every new session, execute once:
Linux:
  $ porter completion bash > /etc/bash_completion.d/porter
MacOS:
  $ porter completion bash > /usr/local/etc/bash_completion.d/porter

You will need to start a new shell for this setup to take effect.
  

```
porter completion bash
```

### Options

```
  -h, --help              help for bash
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

