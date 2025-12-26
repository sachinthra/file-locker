# Contributing to File Locker

Thank you for your interest in contributing to File Locker! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors, regardless of experience level, gender, gender identity and expression, sexual orientation, disability, personal appearance, body size, race, ethnicity, age, religion, or nationality.

### Expected Behavior

- Be respectful and considerate
- Welcome newcomers and help them get started
- Focus on what is best for the project
- Show empathy towards other community members

### Unacceptable Behavior

- Harassment, trolling, or discriminatory language
- Personal attacks or insults
- Publishing others' private information
- Other conduct that could reasonably be considered inappropriate

## Getting Started

### Prerequisites

- **Go:** 1.21 or higher
- **Node.js:** 18 or higher
- **Docker:** 20.10 or higher
- **Git:** 2.30 or higher

### First Time Setup

```bash
# Fork the repository on GitHub

# Clone your fork
git clone https://github.com/YOUR_USERNAME/file-locker.git
cd file-locker

# Add upstream remote
git remote add upstream https://github.com/[original-username]/file-locker.git

# Install development dependencies
make dev-setup
# or
./scripts/setup-dev.sh
```

## Development Setup

### Backend Development

```bash
cd backend

# Install Go dependencies
go mod download

# Run tests
go test ./...

# Run linter
golangci-lint run

# Start development server
go run cmd/server/main.go
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Start development server with hot reload
npm run dev

# Run tests
npm test

# Run linter
npm run lint
```

### Docker Development

```bash
# Build and run all services
docker-compose -f docker-compose.dev.yml up

# View logs
docker-compose logs -f

# Rebuild after changes
docker-compose up --build
```

## How to Contribute

### Reporting Bugs

Before creating a bug report:
1. Check the [issue tracker](https://github.com/[username]/file-locker/issues) for existing issues
2. Update to the latest version to see if the issue persists
3. Collect relevant information (OS, versions, error messages)

When creating a bug report, include:
- **Clear title** describing the issue
- **Steps to reproduce** the problem
- **Expected behavior** vs actual behavior
- **Screenshots** if applicable
- **Environment details** (OS, Go version, Docker version)
- **Error messages or logs**

### Suggesting Features

Feature requests are welcome! When suggesting a feature:
1. Check existing feature requests
2. Provide a clear use case
3. Explain why this feature would be useful
4. Consider how it fits with project goals

### Code Contributions

1. **Find an issue to work on**
   - Look for issues labeled `good-first-issue` or `help-wanted`
   - Comment on the issue to let others know you're working on it

2. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-number-short-description
   ```

3. **Make your changes**
   - Write clean, readable code
   - Follow coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed

4. **Test your changes**
   ```bash
   # Run all tests
   make test
   
   # Run specific tests
   go test ./internal/crypto/...
   npm test -- FileUpload.test.js
   ```

5. **Commit your changes** (see Commit Guidelines)

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request**

## Coding Standards

### Go Code Style

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key points:**
- Use `gofmt` to format code
- Follow Go naming conventions
- Keep functions small and focused
- Write meaningful comments for exported functions
- Handle errors explicitly
- Avoid naked returns

**Example:**
```go
// EncryptFile encrypts a file using AES-256-GCM encryption.
// It returns the encrypted data and any error encountered.
func EncryptFile(plaintext []byte, key []byte) ([]byte, error) {
    if len(key) != 32 {
        return nil, fmt.Errorf("invalid key length: expected 32, got %d", len(key))
    }
    
    // Implementation...
    
    return ciphertext, nil
}
```

### JavaScript/Preact Style

Follow the [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript).

**Key points:**
- Use ESLint with provided configuration
- Use functional components and hooks
- Write descriptive variable names
- Keep components small and single-purpose
- Use PropTypes or TypeScript for type checking

**Example:**
```javascript
// FileUpload.jsx
import { h } from 'preact';
import { useState } from 'preact/hooks';

/**
 * FileUpload component allows users to upload files for encryption.
 * @param {Object} props - Component props
 * @param {Function} props.onUpload - Callback when file is uploaded
 */
export default function FileUpload({ onUpload }) {
    const [isDragging, setIsDragging] = useState(false);
    
    const handleDrop = (e) => {
        e.preventDefault();
        setIsDragging(false);
        // Handle file upload...
    };
    
    return (
        <div className="file-upload">
            {/* Component JSX */}
        </div>
    );
}
```

### Security Guidelines

**Critical rules for security-sensitive code:**

1. **Never log sensitive data**
   ```go
   // ❌ BAD
   log.Printf("User password: %s", password)
   
   // ✅ GOOD
   log.Printf("User authenticated successfully")
   ```

2. **Always zero sensitive memory**
   ```go
   // ✅ GOOD
   defer func() {
       for i := range password {
           password[i] = 0
       }
   }()
   ```

3. **Use constant-time comparisons**
   ```go
   // ✅ GOOD
   if subtle.ConstantTimeCompare(mac1, mac2) == 1 {
       // Valid
   }
   ```

4. **Validate all input**
   ```go
   if len(password) < 8 {
       return ErrPasswordTooShort
   }
   ```

## Commit Guidelines

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat:** New feature
- **fix:** Bug fix
- **docs:** Documentation changes
- **style:** Code style changes (formatting, etc.)
- **refactor:** Code refactoring
- **test:** Adding or updating tests
- **chore:** Build process or auxiliary tool changes

### Examples

```
feat(crypto): implement Argon2id key derivation

Add Argon2id support for secure password-based key derivation
with configurable time, memory, and parallelism parameters.

Closes #42
```

```
fix(grpc): resolve race condition in file upload

Fix race condition that occurred when multiple files were
uploaded simultaneously. Added mutex to protect shared state.

Fixes #123
```

```
docs(readme): update installation instructions

Add detailed steps for Debian package installation and
clarify Docker Compose requirements.
```

## Pull Request Process

1. **Update documentation** if you've changed APIs or behavior

2. **Add tests** for new functionality

3. **Ensure all tests pass**
   ```bash
   make test-all
   ```

4. **Update CHANGELOG.md** if applicable

5. **Fill out the PR template** completely

6. **Request review** from maintainers

7. **Address review comments** promptly

8. **Squash commits** if requested before merge

### PR Title Format

Use the same format as commit messages:
```
feat(crypto): add support for hardware security keys
```

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] All existing tests pass
- [ ] New tests added
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated

## Related Issues
Closes #<issue_number>
```

## Testing

### Unit Tests

```bash
# Go tests
cd backend
go test ./... -v -race -cover

# JavaScript tests
cd frontend
npm test -- --coverage
```

### Integration Tests

```bash
# [To be filled - integration test commands]
make test-integration
```

### Manual Testing

Before submitting a PR, manually test:
1. File encryption and decryption
2. UI responsiveness
3. Error handling
4. Edge cases

### Test Coverage

- Aim for **>80% coverage** for new code
- Critical crypto code should have **>95% coverage**
- UI components should have tests for main flows

## Documentation

### Code Documentation

- **Go:** Use godoc-style comments for exported functions
- **JavaScript:** Use JSDoc comments for functions and components
- **Examples:** Include code examples for complex functions

### User Documentation

When adding features:
1. Update README.md if user-facing
2. Add to ARCHITECTURE.md if technical
3. Create/update relevant guides

### API Documentation

- Update gRPC `.proto` files with comments
- Generate updated documentation: `make docs`

## Questions?

- **General questions:** [GitHub Discussions](https://github.com/[username]/file-locker/discussions)
- **Bug reports:** [GitHub Issues](https://github.com/[username]/file-locker/issues)
- **Security issues:** Email [security email - to be filled]

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes for significant contributions
- Given credit in commit messages and PRs

Thank you for contributing to File Locker! 🔐

---

*Last Updated: December 26, 2025*
