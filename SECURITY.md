# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Currently supported versions:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in tome, please do **not** open a public issue.

### Preferred: GitHub Security Advisories

[Open a private security advisory](https://github.com/kennyg/tome/security/advisories/new)

We will acknowledge your report within 48 hours and provide a more detailed response within 5 business days.

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if available)

## Security Features

### Release Integrity

All tome releases include:

- **SBOM (Software Bill of Materials)**: SPDX-format SBOM for supply chain transparency
- **Signed Artifacts**: Binaries signed with Sigstore/cosign using keyless OIDC
- **Checksums**: SHA-256 checksums for all release artifacts
- **Pinned Dependencies**: GitHub Actions pinned to specific SHAs

### Verification

To verify a release:

```bash
# Download the checksums and signature
wget https://github.com/kennyg/tome/releases/download/v0.x.x/checksums.txt
wget https://github.com/kennyg/tome/releases/download/v0.x.x/checksums.txt.sig
wget https://github.com/kennyg/tome/releases/download/v0.x.x/checksums.txt.pem

# Verify with cosign
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  --certificate-identity-regexp="https://github.com/kennyg/tome" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  checksums.txt

# Verify the binary checksum
sha256sum -c checksums.txt
```

### Continuous Security

- **govulncheck**: Scans for known vulnerabilities in dependencies
- **staticcheck**: Static analysis for Go code quality and security
- **Dependabot**: Automated dependency updates
- **Minimal Permissions**: GitHub Actions use least-privilege permissions

## Scope

We are particularly interested in vulnerabilities related to:

- Command injection or arbitrary code execution
- Path traversal attacks
- Dependency vulnerabilities
- Supply chain attacks
- Authentication/authorization bypasses

## Security Best Practices

When using tome:

1. Always verify release signatures before installation
2. Use the latest version to get security patches
3. Review artifact sources before installation
4. Keep your installation up to date

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security researchers who report valid vulnerabilities.
