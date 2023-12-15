# m3ugen

A CLI tool scanning folders, filtering files and building `M3U` playlists.

## Usage

To install it from this source, use the following script. NB: `$GOPATH/bin` needs to be in your `PATH` variable.

```bash
make install
```

Then call it, specifying at its only argument the configuration file to use.

```bash
m3ugen path/to/configuration_file.yaml
```

Example configuration file (more options are available, but not that relevant to common use, see [config.go](pkg/config.go)):

```yaml
# Path to the output m3u file.
output: example.m3u

# Will display detailed, but not debug information.
# (`debug` shows even more than `verbose`)
verbose: true # Optional. Default: false.
debug: false # Optional. Default: false.

# Will randomize the output list
randomize: true

# Limits the number of entries in the playlist.
maximum: 20

# List of folders to scan
# eg: Will list all files in and under `foo` and `bar`.
scan:
  - ./test_folder_to_scan/foo
  - ./test_folder_to_scan/bar

# List of file extensions to scan for.
# eg: Will output only `*.mp4` and `*.mpg` in the playlist.
extensions:
  - mp4
  - mpg
```

## Development

A useful set of scripts are available through the `make` command.

| `make` Script | Description                                                                        |
| ------------- | ---------------------------------------------------------------------------------- |
| `build`       | Build the main CLI command (`cmd/m3ugen`).                                         |
| `test`        | Run all the automatic tests in the project.                                        |
| `lint`        | Perform `go vet` and additional lint tools to ensure best practices.               |
| `check`       | Good to run before committing. Cleans up the cache, builds, lints, and runs tests. |

Example:

```bash
make check
```
