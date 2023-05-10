## Install

Based on the version in `plugin.yaml`, release binary will be downloaded from GitHub:

```console
$ helm plugin install https://github.com/genchsusu/helm-diff
Downloading and installing helm-diff v0.1.0 ...
https://github.com/genchsusu/helm-diff/releases/download/v0.1.0/helm-diff_0.1.0_darwin_amd64.tar.gz
Installed plugin: diff
```

## Usage

### Show manifest differences

Show manifest differences:

```console
$ helm diff [RELEASE] [CHART] [flags]

Flags:
  -h, --help                     help for diff
  -n, --namespace string         namespace scope for this request (default "orders")
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
  -f, --values strings           specify values in a YAML file or a URL (can specify multiple)
```
