# KAM Version Command


Print out the version of KAM CLI.

```
Usage:
  kam version [flags]

Flags:
  -h, --help   help for version
  ```

```shell
$ kam version
kam version v0.0.5-42-gf2a9373
```

The format of the version string is `<revision-tag>`-`<n-commits>`-g`<commit-sha>`.

```
<revision-tag> - Revision  tag of the release.
<n-commits> - Number of commits after the revision tag has been applied.
<commit-sha> - commit sha 
```
