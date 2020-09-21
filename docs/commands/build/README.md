# Gitops build Command

The `build` sub-command (re-)generates pipelines resources from a pipelines.yaml file.

```shell
$ gitops build
  [--pipelines-folder]
  [--output]
```

| Flag                    | Description |
| ----------------------- | ----------- |
| --help                  | Show help|
| --output                | Optional.  Output path.  (default is the current working directory|
| --pipelines-folder | Optional.  Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml. |

