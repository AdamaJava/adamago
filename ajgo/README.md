# ajgo

ajgo is a commander application written in go that supports the AdamaJava
Java tools available from https://github.com/AdamaJava/adamajava. 

```
Usage:
  ajgo [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  genemodel   Operations on gene models
  genome      Operations on genomes
  gff3        Operations on GFF3 files
  help        Help about any command
  [ngscheck](ngscheck.md)    modes related to proprietary NGScheck tool
  qmotif      Operations related to the AdamaJava qmotif tool
  qpileup     Operations related to the AdamaJava qpileup tool
  seed        Operations on genome seeds
  version     print the ajgo version number

Flags:
      --config string     config file
  -h, --help              help for ajgo
      --logfile string    log file (defaults to STDERR if no file specified)
      --loglevel string   log level (default "INFO")
      --verbose           turn on verbose messaging

Additional help topics:
  ajgo merge-gff3 principles of merging GFF3 files
  ajgo selector   help on selectors and their use

Use "ajgo [command] --help" for more information about a command.
```
