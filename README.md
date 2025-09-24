
# Utility to update the variable or secret in github

Specify `GITHUB_TOKEN` in environment

To Install `go install github.com/git-code11/git-secret-update@latest`

State File Format
NOTE: value and file are mutually exclusive, but value take precedence
```json
[
  {
    key: string,
    value: string,
    file: string,
    secret: bool,
  },
  ...
]
```

Usage
```sh
# Show Help
git-secret-update --help
git-secret-update -repo CompellersHub/lms-react-frontend -file test-state.json
```
