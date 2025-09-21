# Contributing to Praxis Explorer

Thank you for your interest in contributing to Praxis Explorer! This guide will help you get started with developing and contributing to our blockchain AI agent explorer platform.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Code Style and Standards](#code-style-and-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting Guidelines](#issue-reporting-guidelines)
- [Blockchain/Web3 Considerations](#blockchainweb3-considerations)
- [Community Guidelines](#community-guidelines)

## Getting Started

Praxis Explorer is a blockchain explorer specifically designed for AI agents operating on ERC-8004 compatible networks. The project consists of:

- **Backend**: Go-based API service and ERC-8004 indexer
- **Frontend**: Next.js UI for exploring and visualizing agent data
- **Database**: PostgreSQL for storing indexed blockchain data

### Prerequisites

Before contributing, ensure you have the following installed:

- **Docker** (v20.10 or later) and **Docker Compose** (v2.0 or later)
- **Go** (v1.23 or later) for backend development
- **Node.js** (v18 or later) and **npm** for frontend development
- **Git** for version control
- Access to an Ethereum RPC endpoint (for blockchain indexing)

## Development Setup

### 1. Fork and Clone the Repository

```bash
git clone https://github.com/your-username/praxis-explorer.git
cd praxis-explorer
```

### 2. Environment Configuration

Create environment variables for blockchain connectivity:

```bash
# Optional: Set RPC endpoint for on-chain indexing
export SEPOLIA_RPC="https://your-sepolia-rpc-endpoint"
export MAINNET_RPC="https://your-mainnet-rpc-endpoint"
```

### 3. Docker Development Setup

The easiest way to get started is using Docker Compose:

```bash
# Build and start all services
docker compose up --build

# For development with hot reloading
docker compose up --build --watch
```

This will start:
- PostgreSQL database on port 5432
- Backend API on port 8080
- Frontend UI on port 3100

### 4. Local Development Setup

For active development, you may prefer running services locally:

#### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Set up database connection
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/praxis_explorer?sslmode=disable"
export ERC8004_CONFIG="./configs/erc8004.yaml"

# Run the backend
go run cmd/praxis-explorer/main.go
```

#### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Set environment variables
export NEXT_PUBLIC_API_URL="http://localhost:8080"

# Start development server
npm run dev
```

## Project Structure

```
praxis-explorer/
├── backend/                  # Go backend service
│   ├── cmd/                 # Main applications
│   ├── internal/            # Private application code
│   │   ├── erc8004/        # ERC-8004 smart contract integration
│   │   ├── explorer/       # Core explorer logic
│   │   └── store/          # Database models and operations
│   ├── migrations/         # Database migrations
│   └── configs/           # Configuration files
├── frontend/               # Next.js frontend
│   ├── app/               # Next.js 13+ app directory
│   ├── components/        # React components
│   ├── hooks/            # Custom React hooks
│   └── types/            # TypeScript type definitions
└── docker-compose.yml    # Development environment
```

## Development Workflow

### Branch Naming Convention

Use descriptive branch names with the following prefixes:
- `feature/` - New features
- `bugfix/` - Bug fixes
- `hotfix/` - Critical fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates

Example: `feature/add-agent-reputation-display`

### Commit Message Format

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

Examples:
```
feat(explorer): add agent reputation scoring system
fix(indexer): resolve block range query optimization
docs(api): update swagger documentation for agent endpoints
```

## Code Style and Standards

### Go Backend Standards

- Follow standard Go conventions and use `gofmt`
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Handle errors explicitly and provide meaningful error messages
- Use dependency injection for better testability

Example:
```go
// GetAgentByID retrieves agent information by ID from the blockchain
func (s *Store) GetAgentByID(ctx context.Context, agentID *big.Int) (*models.Agent, error) {
    if agentID == nil {
        return nil, fmt.Errorf("agent ID cannot be nil")
    }
    // Implementation...
}
```

### Frontend Standards

- Use TypeScript for all new code
- Follow React best practices and hooks patterns
- Use Tailwind CSS for styling
- Implement proper error handling and loading states
- Use SWR for data fetching and caching

Example:
```typescript
interface AgentCardProps {
  agent: Agent;
  onSelect?: (agent: Agent) => void;
}

export const AgentCard: React.FC<AgentCardProps> = ({ agent, onSelect }) => {
  // Component implementation...
};
```

### Blockchain Integration Standards

- Always validate blockchain data before processing
- Implement proper retry logic for RPC calls
- Use appropriate gas estimation for transactions
- Handle network-specific configurations
- Implement proper event filtering and indexing

## Testing Requirements

### Backend Testing

Run tests using:
```bash
cd backend
go test ./...
```

Requirements:
- Unit tests for all business logic
- Integration tests for database operations
- Mock external dependencies (RPC calls, etc.)
- Test error conditions and edge cases

### Frontend Testing

```bash
cd frontend
npm test
```

Requirements:
- Component tests using React Testing Library
- Integration tests for API interactions
- E2E tests for critical user flows
- Accessibility testing

### Blockchain Testing

- Test against local test networks when possible
- Use mock contracts for unit testing
- Validate ABI compatibility
- Test event parsing and indexing logic

## Pull Request Process

### Before Submitting

1. **Ensure all tests pass**: Run the full test suite
2. **Update documentation**: Update relevant docs for your changes
3. **Check code quality**: Run linters and formatters
4. **Test manually**: Verify your changes work as expected
5. **Review security**: Especially for blockchain-related changes

### PR Requirements

1. **Clear description**: Explain what the PR does and why
2. **Link related issues**: Reference any related GitHub issues
3. **Screenshots/demos**: For UI changes, include visual evidence
4. **Breaking changes**: Clearly document any breaking changes
5. **Security considerations**: Note any security implications

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added new tests for new functionality
- [ ] Manual testing completed

## Blockchain Considerations
- [ ] RPC compatibility verified
- [ ] Gas usage optimized
- [ ] Event parsing tested
- [ ] Network configuration updated

## Screenshots
(If applicable)

## Additional Notes
Any additional information
```

## Issue Reporting Guidelines

### Bug Reports

Use the bug report template and include:
- **Environment details**: OS, browser, network
- **Steps to reproduce**: Clear, numbered steps
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Screenshots/logs**: Visual evidence or error logs
- **Blockchain context**: Network, block number, transaction hash

### Feature Requests

Include:
- **Problem statement**: What problem does this solve?
- **Proposed solution**: How should it work?
- **Alternatives considered**: Other approaches you've thought of
- **Additional context**: Why is this important?

### Security Issues

**Do not report security issues publicly.** Instead:
1. Email security concerns to [security@praxis.example] (placeholder)
2. Include detailed information about the vulnerability
3. Allow reasonable time for fixes before disclosure

## Blockchain/Web3 Considerations

### ERC-8004 Compliance

- Follow ERC-8004 specifications for agent identity
- Validate agent signatures properly
- Implement reputation scoring according to standard
- Handle identity resolution correctly

### Network Support

- Test on testnets before mainnet
- Handle network-specific configurations
- Implement proper error handling for network issues
- Consider gas costs and optimization

### Security Best Practices

- Validate all inputs from blockchain
- Use proper cryptographic libraries
- Implement rate limiting for RPC calls
- Handle private key security (if applicable)
- Audit smart contract interactions

### Performance Considerations

- Implement efficient block range queries
- Use appropriate indexing strategies
- Cache frequently accessed data
- Handle large datasets properly

## Community Guidelines

### Code of Conduct

All contributors must follow our [Code of Conduct](CODE_OF_CONDUCT.md). We are committed to providing a welcoming and inclusive environment for all contributors.

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Pull Requests**: Code contributions and reviews

### Getting Help

If you're stuck or need guidance:

1. Check existing documentation and issues
2. Ask questions in GitHub Discussions
3. Reach out to maintainers in PR comments
4. Join our community channels (if available)

## Additional Resources

- [ERC-8004 Specification](https://example.com/erc8004) (placeholder)
- [Ethereum Development Documentation](https://ethereum.org/en/developers/)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [React Documentation](https://reactjs.org/docs/)
- [Next.js Documentation](https://nextjs.org/docs)

## License

By contributing to Praxis Explorer, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Praxis Explorer! Together, we're building the future of decentralized AI agent exploration and interaction.