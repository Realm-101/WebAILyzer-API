# Contributing to WebAIlyzer Lite API

Thank you for your interest in contributing to the WebAIlyzer Lite API! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and professional in all interactions.

### Our Standards

- Use welcoming and inclusive language
- Be respectful of differing viewpoints and experiences
- Gracefully accept constructive criticism
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- Go 1.21 or later installed
- Docker and Docker Compose for local development
- Git for version control
- A GitHub account for submitting pull requests

### Development Setup

1. **Fork the repository**
   ```bash
   # Fork the repo on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/wappalyzergo.git
   cd wappalyzergo
   ```

2. **Set up the development environment**
   ```bash
   # Install dependencies
   make deps
   
   # Start development services
   docker-compose up -d postgres redis
   
   # Run database migrations
   make migrate-up
   ```

3. **Verify the setup**
   ```bash
   # Run tests to ensure everything works
   make test
   
   # Start the API server
   make run
   ```

## Contributing Process

### 1. Choose an Issue

- Look for issues labeled `good first issue` for beginners
- Check existing issues before creating new ones
- Comment on issues you'd like to work on

### 2. Create a Branch

```bash
# Create a feature branch from main
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### 3. Make Changes

- Follow the coding standards outlined below
- Write tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 4. Commit Changes

```bash
# Stage your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add new analysis endpoint for batch processing"
```

#### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(analysis): add accessibility analyzer
fix(auth): resolve token validation issue
docs: update API documentation
test: add integration tests for metrics endpoint
```

## Coding Standards

### Go Code Style

We follow standard Go conventions and use automated tools to enforce consistency:

```bash
# Format code
make fmt

# Run linter
make lint

# Check for common issues
go vet ./...
```

### Code Organization

- **Packages**: Use clear, descriptive package names
- **Functions**: Keep functions small and focused
- **Interfaces**: Define interfaces in the package that uses them
- **Error Handling**: Always handle errors explicitly
- **Comments**: Write clear comments for exported functions and types

### Example Code Structure

```go
// Package analysis provides website analysis capabilities.
package analysis

import (
    "context"
    "fmt"
    
    "github.com/projectdiscovery/wappalyzergo/internal/models"
)

// Analyzer defines the interface for website analysis.
type Analyzer interface {
    Analyze(ctx context.Context, url string) (*models.AnalysisResult, error)
}

// Service implements the Analyzer interface.
type Service struct {
    repo Repository
    logger Logger
}

// NewService creates a new analysis service.
func NewService(repo Repository, logger Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}

// Analyze performs website analysis for the given URL.
func (s *Service) Analyze(ctx context.Context, url string) (*models.AnalysisResult, error) {
    if url == "" {
        return nil, fmt.Errorf("url cannot be empty")
    }
    
    // Implementation here...
    
    return result, nil
}
```

### Database Guidelines

- Use migrations for all schema changes
- Write efficient queries with proper indexing
- Use transactions for multi-step operations
- Handle database errors gracefully

### API Guidelines

- Follow RESTful conventions
- Use appropriate HTTP status codes
- Validate all input parameters
- Return consistent error responses
- Document all endpoints

## Testing Guidelines

### Test Structure

We use a comprehensive testing strategy:

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test database operations and external services
3. **End-to-End Tests**: Test complete API workflows

### Writing Tests

```go
func TestAnalysisService_Analyze(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        want    *models.AnalysisResult
        wantErr bool
    }{
        {
            name: "valid URL",
            url:  "https://example.com",
            want: &models.AnalysisResult{
                URL: "https://example.com",
                // ... expected result
            },
            wantErr: false,
        },
        {
            name:    "empty URL",
            url:     "",
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := &Service{} // Initialize service
            got, err := s.Analyze(context.Background(), tt.url)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Analyze() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Analyze() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run benchmarks
make test-benchmarks
```

### Test Requirements

- All new code must have tests
- Maintain or improve test coverage
- Tests must be deterministic and fast
- Use table-driven tests for multiple scenarios
- Mock external dependencies

## Documentation

### Code Documentation

- Document all exported functions and types
- Use clear, concise comments
- Include examples for complex functionality
- Keep documentation up to date with code changes

### API Documentation

- Update OpenAPI specifications for API changes
- Include request/response examples
- Document error conditions
- Update integration examples

### README Updates

Update the README.md file when:
- Adding new features
- Changing configuration options
- Modifying setup instructions
- Adding new dependencies

## Submitting Changes

### Pull Request Process

1. **Ensure your branch is up to date**
   ```bash
   git checkout main
   git pull upstream main
   git checkout your-feature-branch
   git rebase main
   ```

2. **Run the full test suite**
   ```bash
   make test-all
   ```

3. **Push your changes**
   ```bash
   git push origin your-feature-branch
   ```

4. **Create a Pull Request**
   - Use a clear, descriptive title
   - Reference related issues
   - Describe the changes made
   - Include testing instructions

### Pull Request Template

```markdown
## Description
Brief description of the changes made.

## Related Issues
Fixes #123
Relates to #456

## Changes Made
- Added new analysis endpoint
- Updated documentation
- Added integration tests

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and linting
2. **Code Review**: Maintainers review the code for quality and correctness
3. **Feedback**: Address any feedback or requested changes
4. **Approval**: Once approved, the PR will be merged

## Getting Help

### Communication Channels

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Pull Request Comments**: For code-specific discussions

### Resources

- [Go Documentation](https://golang.org/doc/)
- [Project Documentation](./API_DOCUMENTATION.md)
- [Troubleshooting Guide](./TROUBLESHOOTING.md)

## Recognition

Contributors will be recognized in:
- The project's contributor list
- Release notes for significant contributions
- The project README for major features

Thank you for contributing to WebAIlyzer Lite API! Your contributions help make this project better for everyone.