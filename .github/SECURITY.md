# Security Policy

## Supported Versions

We actively support the following versions of Praxis Explorer with security updates:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < Latest| :x:                |

## Reporting a Vulnerability

We take the security of Praxis Explorer seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities via one of the following methods:

1. **Email**: Send details to [security@praxis.dev] (replace with actual security contact)
2. **GitHub Security Advisories**: Use the "Report a vulnerability" button in the Security tab of this repository

### What to Include

Please include the following information in your report:

- **Type of issue** (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- **Full paths of source file(s) related to the manifestation of the issue**
- **The location of the affected source code** (tag/branch/commit or direct URL)
- **Any special configuration required to reproduce the issue**
- **Step-by-step instructions to reproduce the issue**
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the issue**, including how an attacker might exploit it

### Response Timeline

- **Initial Response**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Assessment**: We will assess the vulnerability and provide an initial response within 5 business days
- **Updates**: We will keep you informed of our progress throughout the investigation
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days

## Security Considerations

### Blockchain-Specific Security

Given that Praxis Explorer interacts with blockchain networks and smart contracts, please pay special attention to:

- **Smart Contract Interactions**: Potential for reentrancy attacks, integer overflow/underflow
- **Private Key Management**: Secure handling of cryptographic materials
- **RPC Endpoint Security**: Protection against malicious RPC responses
- **Agent Registry Security**: Validation of agent registrations and signatures
- **ERC-8004 Implementation**: Compliance with the standard and secure signature verification

### Web Application Security

Common web application vulnerabilities to consider:

- **Input Validation**: SQL injection, XSS, command injection
- **Authentication & Authorization**: Privilege escalation, session management
- **Data Exposure**: Sensitive information leakage
- **API Security**: Rate limiting, input validation, output encoding

### Infrastructure Security

- **Docker Security**: Container escape, privilege escalation
- **Database Security**: Injection attacks, unauthorized access
- **Network Security**: Man-in-the-middle attacks, protocol vulnerabilities

## Disclosure Policy

When we receive a security bug report, we will:

1. **Confirm the problem** and determine the affected versions
2. **Audit code** to find any potential similar problems
3. **Prepare fixes** for all supported versions
4. **Release new versions** as soon as possible
5. **Announce the vulnerability** in a security advisory

## Security Best Practices

### For Contributors

- **Code Review**: All code changes must be reviewed by at least one other developer
- **Dependency Management**: Regularly update dependencies and monitor for vulnerabilities
- **Static Analysis**: Use static analysis tools to identify potential security issues
- **Testing**: Include security-focused test cases
- **Secrets Management**: Never commit secrets, API keys, or private keys to the repository

### For Users

- **Keep Updated**: Always use the latest version of Praxis Explorer
- **Secure Configuration**: Follow security best practices for Docker, database, and network configuration
- **Monitor Logs**: Regularly monitor application and system logs for suspicious activity
- **Access Control**: Implement proper access controls for your deployment
- **Backup Strategy**: Maintain secure backups of your data

## Known Security Considerations

### Current Security Measures

- **Input Validation**: All API inputs are validated and sanitized
- **Database Security**: Parameterized queries used to prevent SQL injection
- **CORS Configuration**: Proper CORS settings configured for frontend-backend communication
- **Container Security**: Non-root user execution in Docker containers

### Areas for Ongoing Attention

- **Rate Limiting**: Consider implementing rate limiting for API endpoints
- **Audit Logging**: Enhanced logging for security-relevant events
- **Cryptographic Libraries**: Regular updates to cryptographic dependencies
- **Network Security**: TLS configuration and certificate management

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Ethereum Smart Contract Security Best Practices](https://consensys.github.io/smart-contract-best-practices/)
- [Go Security Checklist](https://github.com/securego/gosec)
- [Next.js Security Guidelines](https://nextjs.org/docs/advanced-features/security-headers)

## Contact

For questions about this security policy, please contact:
- **General Security Questions**: [security@praxis.dev] (replace with actual contact)
- **Project Maintainers**: [See MAINTAINERS.md or package.json]

---

**Note**: This security policy is a living document and will be updated as the project evolves.