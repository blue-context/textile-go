# Contributing to Textile-Go

Thank you for your interest in contributing to Textile-Go! We welcome contributions from the community.

## Development Setup

### Prerequisites

- Go 1.24.5 or later
- Git

### Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/textile-go.git
   cd textile-go
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

## Code Standards

### Go Style Guide

Follow these standards for all contributions:

- **Go idioms**: Follow [Effective Go](https://go.dev/doc/effective_go) and [Google Go Style Guide](https://google.github.io/styleguide/go/)
- **Error handling**: Always handle errors explicitly, never ignore them
- **Context usage**: Use `context.Context` as first parameter for cancellation/timeout
- **Documentation**: Document all exported types, functions, and methods
- **Testing**: Write table-driven tests with subtests
- **Concurrency**: Ensure thread-safety for concurrent operations

### Code Quality Requirements

All contributions must meet these requirements:

- ✅ Complete implementations (no TODOs, FIXMEs, or stubs)
- ✅ Proper error handling for all edge cases
- ✅ Tests pass with race detector: `go test -race ./...`
- ✅ Code coverage >80% for new features
- ✅ No ignored errors (avoid using `_` for error returns)
- ✅ Clean, maintainable code following established patterns

### File Headers

All Go source files must include the Apache 2.0 license header:

```go
// Copyright 2025 Blue Context Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
```

## Running Tests

### Run all tests with race detector:
```bash
go test -race ./...
```

### Run tests with coverage:
```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run benchmarks:
```bash
go test -bench=. -benchmem ./...
```

## Making Changes

### Workflow

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the code standards above

3. Write/update tests for your changes

4. Run tests and ensure they pass:
   ```bash
   go test -race ./...
   ```

5. Commit your changes using [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: add new transformer type"
   git commit -m "fix: resolve context cancellation issue"
   git commit -m "test: add table-driven tests for pipeline"
   ```

6. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

7. Open a Pull Request

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring (no behavior change)
- `test`: Adding or updating tests
- `docs`: Documentation changes
- `perf`: Performance improvements
- `style`: Code style/formatting changes

**Examples:**
```
feat(pipeline): add transformer composition support
fix(streaming): resolve race condition in chunk processing
test(client): add coverage for error strategies
docs(readme): update installation instructions
```

## Pull Request Guidelines

### Before Submitting

- [ ] All tests pass with race detector
- [ ] Code coverage maintained or improved
- [ ] Documentation updated (if needed)
- [ ] Conventional commit messages used
- [ ] No TODOs, FIXMEs, or placeholder code
- [ ] License headers added to all new Go files

### PR Description

Include in your PR description:

1. **What**: Brief description of the change
2. **Why**: Motivation for the change
3. **How**: High-level implementation approach
4. **Testing**: How you tested the changes
5. **Breaking Changes**: Any backwards-incompatible changes (if applicable)

## Code Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR

## Questions or Issues?

- Open an issue for bugs or feature requests
- Use discussions for questions or ideas

## License

By contributing to Textile-Go, you agree that your contributions will be licensed under the Apache License 2.0.

Copyright 2025 Blue Context Inc.
