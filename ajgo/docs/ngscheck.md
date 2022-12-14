# ngscheck

NGScheck is a proprietary quality control tool for assessing
next-generation sequencing. NGScheck was created by genomiQa and
operates on the outputs of The AdamnaJava/adamajava qprofiler2
tool which is why these modes are included here. The submodes under
ngscheck all operate on the outputs of NGScheck so they are only
useful if you have an NGScheck license.

```
Usage:
  ajgo ngscheck [command]

Available Commands:
  collect     collect NGScheck info for a list of BAM files
  debug       check that .ngcbas.json file can be parsed

Flags:
  -h, --help   help for ngscheck

Global Flags:
      --config string     config file
      --logfile string    log file (defaults to STDERR if no file specified)
      --loglevel string   log level (default "INFO")
      --verbose           turn on verbose messaging

Use "ajgo ngscheck [command] --help" for more information about a command.
```
