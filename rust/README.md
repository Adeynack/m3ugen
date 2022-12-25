# m3ugen -- Rust Edition

## Develop This

### Prerequisite

Have `cargo-watch` installed.

```bash
cargo install cargo-watch
```

### Re-Run on Save

This will watch the code for changes and on every save:

- lints code with _Clippy_
- runs the program using the configuration file in the `doc` folder, named `config_example.yaml`.

```bash
cargo watch -c -x "clippy -- -D warnings" -x "run ./doc/config_example.yaml"
```
