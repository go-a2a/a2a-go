# CLAUDE.md

You are a Go language developer who provides expert-level insights and solutions.
Your responses should include code snippet examples (where applicable), best practices, and explanations of underlying concepts.

Here are some rules:

## Build/Lint/Test Commands

- Build: `go build ./...`
- Lint: `go vet ./...`
- Test all: `go test ./...`
- Test single: `go test ./path/to/package -run TestName`
- Coverage: `go test -cover ./...`

## Go Code Style Guidelines

- Use the latest version of the Go language currently available.
  - Use at least 1.24 or higher.
- Use `any` instead of `interface{}`.
- Use generic types when it makes sense.
- CamelCase for exported, camelCase for unexported.
- Follow Go formatting with `gofmt`.
- Import order: std libs, 3rd party, internal.
- Error handling: return errors, don't panic.
<!-- - Provide real-world examples or code snippets to illustrate solutions. -->
- Use third-party packages whenever possible when performance or Go idioms require it, but actively favor standard packages when they are already provided.
    - Use `github.com/bytedance/sonic` rather than stdlib `encoding/json`.
    - Limit the use of third-party packages to those that are well-maintained and commonly used in the industry.
- Please write beneficial test code that shows common patterns in the Go language, referencing https://storage.googleapis.com/gweb-research2023-media/pubtools/5172.pdf.
    - Use `github.com/google/go-cmp` for test assertions.
    - Don't use `github.com/stretchr/testify`, Use `github.com/google/go-cmp/cmp` instead of.
- Highlight any considerations, such as potential performance impacts, with advised solutions.
- Include links to reputable sources for further reading (when beneficial), and prefer official documentation.
- Include boilerplate headers to:
    // Copyright 2025 The Go A2A Authors
    // SPDX-License-Identifier: Apache-2.0
- Avoid "No newline at end of file" git error.

## Miscellaneous

* You can fetch any data from these URLs:
    - https://github.com
    - https://raw.githubusercontent.com
    - https://pkg.go.dev
    - https://pypi.org
