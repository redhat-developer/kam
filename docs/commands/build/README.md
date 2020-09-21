# Gitops build Command

The `build` sub-command (re-)generates pipelines resources from a pipelines.yaml file.

```shell
$ gitops build
  [--pipelines-file]
  [--output]
```

| Flag                    | Description |
| ----------------------- | ----------- |
| --help                  | Show help|
| --output                | Optional.  Output path.  (default is the current working directory|
| --pipelines-file | Optional.  Path to pipelines file.  Default is _pipelines.yaml_. |

