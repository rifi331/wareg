# Contributing to Cooking App

Thank you for your interest in contributing to the Cooking App! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and inclusive in all interactions.

## How to Contribute

### Reporting Bugs

1. Check existing [issues](../../issues)
2. Create a new issue with:
   - Clear title
   - Detailed description
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Check existing feature requests
2. Create a new issue with:
   - Feature description
   - Use case/problem it solves
   - Alternative solutions considered

### Submitting Code Changes

1. Fork the repository
2. Create a feature branch:

```bash
git checkout -b feature/your-feature-name
```

3. Make your changes following coding standards
4. Test thoroughly
5. Commit with clear messages:

```bash
git commit -m "Add: feature description"
# or
git commit -m "Fix: bug description"
```

6. Push to your fork:

```bash
git push origin feature/your-feature-name
```

7. Create a Pull Request with:
   - Clear title
   - Description of changes
   - Related issue numbers

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` to format code
- Write clear, concise functions
- Avoid unnecessary comments
- Add error handling

### Frontend

- Mobile-first responsive design
- Semantic HTML
- Accessible components
- Clean HTMX patterns

### Database

- Write idempotent migration scripts
- Use transactions for multi-step operations
- Index frequently queried columns
- Document schema changes

## Development Workflow

### Setup Development Environment

1. Clone your fork:

```bash
git clone https://github.com/your-username/wareg.git
cd wareg
```

2. Install dependencies:

```bash
go mod download
```

3. Set up database and run schema

4. Copy environment file:

```bash
cp .env.example .env
# Edit .env with your settings
```

5. Run application:

```bash
go run main.go
```

### Testing

Before submitting:

1. Test all modified features
2. Check for regressions
3. Test on mobile devices (responsive)
4. Verify database operations

### Pull Request Guidelines

- Keep PRs focused and small
- Update documentation if needed
- Add tests for new features
- Ensure CI/CD passes
- Respond to review comments promptly

## Project Structure

```
wareg/
├── .github/workflows/     # CI/CD configurations
├── helm/                  # Helm chart for deployment
├── sql/                   # Database schema and migrations
├── frontend/templates/     # HTML templates with HTMX
├── static/                # Static assets
├── main.go               # Go application entry point
├── go.mod               # Go module dependencies
├── Dockerfile           # Container definition
├── docker-compose.yml   # Local development
└── README.md           # Documentation
```

## Getting Help

- Create an issue for bugs or questions
- Check existing issues and discussions
- Refer to documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
