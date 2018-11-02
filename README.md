# dumpstructs
Utility to dump the structs from Golang sources into a unified view

# Usage

`dumpstructs` may be run without any arguments and it will traveserse your
current working directory for any `.go` files, then output all their structs.

Here are the full options though, which feel pretty self-explanatory:

```
NAME:
   dumpstructs

USAGE:
   dumpstructs [options]

VERSION:
   0.0.1

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path value, -p value     Path to traverse to discover Go files. Optional. (default: ".")
   --exclude value, -x value  Regex pattern to use to exclude paths from list. Optional.
   --help, -h                 show help
   --version, -v              print the version
```
