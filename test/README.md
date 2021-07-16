# Integration tests

This tests are run against a deployed node using the Go SDK. It's assumed that the node is working well before the test starts.

## Usage

1. Before runing a test, read the requirements specified at the beginning of `<test name>/main.go`.
2. Copy `test/config.example.toml` into `test/config.toml` and edit the values, making sure that the requirements are fulfilled.
3. Run the test: `cd test && go run ./<test name>`
