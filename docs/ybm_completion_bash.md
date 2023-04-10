## ybm completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(ybm completion bash)

To load completions for every new session, execute once:

#### Linux:

	ybm completion bash > /etc/bash_completion.d/ybm

#### macOS:

	ybm completion bash > $(brew --prefix)/etc/bash_completion.d/ybm

You will need to start a new shell for this setup to take effect.


```
ybm completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -a, --apiKey string     YBM Api Key
      --config string     config file (default is $HOME/.ybm-cli.yaml)
      --debug             Use debug mode, same as --logLevel debug
  -l, --logLevel string   Select the desired log level format(info). Default to info
      --no-color          Disable colors in output , default to false
  -o, --output string     Select the desired output format (table, json, pretty). Default to table
      --wait              Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm completion](ybm_completion.md)	 - Generate the autocompletion script for the specified shell

