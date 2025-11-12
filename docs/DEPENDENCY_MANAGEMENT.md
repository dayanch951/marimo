# Dependency Management Guide

Guide for managing dependencies in Marimo ERP project.

## Table of Contents

1. [Go Dependencies](#go-dependencies)
2. [Frontend Dependencies](#frontend-dependencies)
3. [Security Updates](#security-updates)
4. [Automated Updates](#automated-updates)
5. [Troubleshooting](#troubleshooting)

## Go Dependencies

### Checking for Updates

```bash
# List all modules and their versions
go list -m all

# Check for available updates
go list -u -m all

# Check specific module
go list -m -u github.com/gin-gonic/gin
```

### Updating Dependencies

```bash
# Update all dependencies to latest minor/patch versions
go get -u ./...

# Update to latest major versions (may break compatibility)
go get -u=patch ./...  # Only patch updates
go get -u=minor ./...  # Minor and patch updates

# Update specific package
go get -u github.com/gin-gonic/gin

# Clean up unused dependencies
go mod tidy

# Verify checksums
go mod verify
```

### Adding New Dependencies

```bash
# Add dependency
go get github.com/pkg/errors@latest

# Add specific version
go get github.com/pkg/errors@v0.9.1

# Add and update go.mod
go get github.com/pkg/errors
go mod tidy
```

### Removing Dependencies

```bash
# Remove from code, then:
go mod tidy
```

### Version Constraints

```go
// go.mod
module marimo

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/google/uuid v1.5.0
    gorm.io/gorm v1.25.5
)
```

## Frontend Dependencies

### Checking for Updates

```bash
cd frontend

# Check outdated packages
npm outdated

# Check for major version updates
npm outdated --long
```

### Updating Dependencies

```bash
# Update all packages (respecting semver in package.json)
npm update

# Update to latest versions (may break)
npm install package@latest

# Update specific package
npm install react@latest

# Interactive update
npx npm-check-updates -i
```

### Adding New Dependencies

```bash
# Add production dependency
npm install axios

# Add dev dependency
npm install --save-dev @types/node

# Add specific version
npm install axios@1.6.0
```

### Removing Dependencies

```bash
# Remove package
npm uninstall axios

# Remove dev package
npm uninstall --save-dev @types/node
```

### Security Auditing

```bash
# Run security audit
npm audit

# Fix vulnerabilities automatically
npm audit fix

# Fix including breaking changes
npm audit fix --force
```

## Security Updates

### Go Security

```bash
# Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Or use govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### npm Security

```bash
# Security audit
npm audit

# Fix automatically
npm audit fix

# Review audit report
npm audit --json

# Check specific package
npm view package-name
```

### Snyk Integration

```bash
# Install Snyk CLI
npm install -g snyk

# Authenticate
snyk auth

# Test for vulnerabilities
snyk test

# Monitor project
snyk monitor
```

## Automated Updates

### Dependabot Configuration

Create `.github/dependabot.yml`:

```yaml
version: 2
updates:
  # Go modules
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10

  # Frontend npm
  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10

  # Mobile npm
  - package-ecosystem: "npm"
    directory: "/mobile"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10

  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

### Update Script

Use the automated update script:

```bash
# Run update script
./scripts/update-dependencies.sh
```

This script:
1. Updates Go dependencies
2. Updates frontend dependencies
3. Updates mobile dependencies
4. Runs tests
5. Shows summary

### CI/CD Integration

Dependencies are checked in CI pipeline:

```yaml
# .github/workflows/ci.yml
- name: Check for outdated dependencies
  run: |
    go list -u -m all | grep '\['
    cd frontend && npm outdated

- name: Security audit
  run: |
    govulncheck ./...
    cd frontend && npm audit
```

## Update Strategy

### When to Update

- **Security patches**: Immediately
- **Bug fixes**: Within 1 week
- **Minor versions**: Monthly
- **Major versions**: Quarterly (after testing)

### Update Process

1. **Check updates**
   ```bash
   ./scripts/update-dependencies.sh --dry-run
   ```

2. **Review changelog**
   - Check breaking changes
   - Review migration guides
   - Assess impact

3. **Update in branch**
   ```bash
   git checkout -b chore/update-dependencies
   ./scripts/update-dependencies.sh
   ```

4. **Test thoroughly**
   ```bash
   # Run all tests
   make test

   # Manual testing
   # - Critical user flows
   # - Integration points
   # - Performance benchmarks
   ```

5. **Update documentation**
   - Update version requirements
   - Document breaking changes
   - Update setup instructions

6. **Create PR**
   ```bash
   git add go.mod go.sum frontend/package.json frontend/package-lock.json
   git commit -m "chore: update dependencies"
   git push origin chore/update-dependencies
   ```

7. **Code review**
   - Review dependency changes
   - Check for security issues
   - Verify tests pass

8. **Deploy to staging**
   - Test in staging environment
   - Monitor for issues
   - Rollback if needed

9. **Deploy to production**
   - Deploy during low-traffic period
   - Monitor closely
   - Have rollback plan ready

## Version Pinning

### When to Pin Versions

Pin exact versions when:
- Production stability is critical
- Known issues with newer versions
- Compatibility requirements

### How to Pin

**Go**:
```bash
# Pin to specific version
go get github.com/pkg/errors@v0.9.1

# go.mod will have:
require github.com/pkg/errors v0.9.1
```

**npm**:
```json
{
  "dependencies": {
    "react": "18.2.0"  // Exact version (no ^ or ~)
  }
}
```

## Troubleshooting

### Go Module Issues

**Problem**: Dependency resolution errors

```bash
# Clear module cache
go clean -modcache

# Re-download modules
go mod download

# Verify modules
go mod verify
```

**Problem**: Incompatible versions

```bash
# Check why package is needed
go mod why github.com/pkg/errors

# Find specific version
go get github.com/pkg/errors@v0.9.1
```

### npm Issues

**Problem**: Package lock conflicts

```bash
# Delete lock file and reinstall
rm package-lock.json
npm install
```

**Problem**: Peer dependency errors

```bash
# Install with legacy peer deps
npm install --legacy-peer-deps
```

**Problem**: Cache issues

```bash
# Clear npm cache
npm cache clean --force

# Verify cache
npm cache verify
```

### Common Issues

#### "incompatible with go.mod"

```bash
# Update go version in go.mod
go 1.21  # Update to match installed version

# Or update Go installation
```

#### "ERESOLVE unable to resolve dependency tree"

```bash
# Use --force or --legacy-peer-deps
npm install --legacy-peer-deps
```

#### "checksum mismatch"

```bash
# Go
go clean -modcache
go mod download

# npm
rm -rf node_modules package-lock.json
npm install
```

## Best Practices

1. **Regular Updates**
   - Schedule weekly dependency checks
   - Update minor versions monthly
   - Major versions quarterly

2. **Testing**
   - Always run full test suite
   - Test in staging before production
   - Have rollback plan

3. **Documentation**
   - Document breaking changes
   - Update changelog
   - Notify team

4. **Security**
   - Subscribe to security advisories
   - Run security audits weekly
   - Prioritize security patches

5. **Review**
   - Read changelogs before updating
   - Check for breaking changes
   - Review dependency tree

## Monitoring

### Dependency Health

Monitor:
- Outdated packages (weekly)
- Security vulnerabilities (daily)
- Deprecated packages (monthly)
- License compliance (quarterly)

### Tools

- **Dependabot**: Automated PRs for updates
- **Snyk**: Security monitoring
- **npm audit**: Security vulnerabilities
- **govulncheck**: Go vulnerabilities
- **GitHub Security Alerts**: Automatic alerts

## Resources

- [Go Modules Reference](https://go.dev/ref/mod)
- [npm Documentation](https://docs.npmjs.com/)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
