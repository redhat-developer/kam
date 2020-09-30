# KAM build Command

The `build` sub-command (re-)generates pipelines resources from a pipelines.yaml file.

```
Build GitOps pipelines files

Usage:
  kam build [flags]

Examples:
  # Build files from pipelines
  kam build

Flags:
  -h, --help                      help for build
      --output string             Folder path to add GitOps resources (default ".")
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
```
```shell
$ kam build
  [--pipelines-folder]
  [--output]
```
