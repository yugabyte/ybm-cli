## ybm completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	ybm completion fish | source

To load completions for every new session, execute once:

	ybm completion fish > ~/.config/fish/completions/ybm.fish

You will need to start a new shell for this setup to take effect.


```
ybm completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm completion](ybm_completion.md)	 - Generate the autocompletion script for the specified shell

