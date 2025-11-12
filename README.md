# Recs Votes Storage

A production-grade REST API service for managing user votes and relationships ("romances") in Bumble 2.0's recommendation system. This service handles voting interactions between users, tracks vote statistics (hourly and lifetime), and manages asynchronous cleanup operations for bulk deletions. oi

## Architecture & Design

- **System Design**: [User Votes Storage - System Design Bumble 2.0](https://bmbl.atlassian.net/wiki/spaces/REC/pages/3082616846/User+Votes+Storage+-+System+Design+Bumble+2.0)
- **AWS Architecture**: [Bumble 2.0 Implementation on AWS](https://bmbl.atlassian.net/wiki/spaces/REC/pages/3215360091/Bumble+2.0+Implementation+on+AWS+Medium-level+Draft)
- **System Design Diagram**: [Figma Board](https://www.figma.com/board/pamePszOjJFv8vtyn17ztt/Bumble-2.0---Recommendations-and-Voting?node-id=16-268&p=f&t=CwVzMCs6KCbEq0n0-0)

## Prerequisites

Before you begin, ensure you have the following installed:

### Required Tools

```bash
# Install Go
brew install go

# Install AWS CDK globally
npm install -g aws-cdk aws-cdk-local

# Install Wire (dependency injection code generator)
go install github.com/google/wire/cmd/wire@latest

# Install Lefthook (git hooks)
go install github.com/evilmartians/lefthook@latest

# Install Conform (for commit message validation)
go install github.com/siderolabs/conform/cmd/conform@latest

# Install hooks
lefthook install

# Add Go bin to PATH (if not already in your shell profile)
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Verify Installation

```bash
go version        # Should be Go 1.25 or higher
cdk --version     # Should be AWS CDK 2.x
cdklocal --version
wire help    # If command not found, ensure Go bin is in your PATH
go version -m "$(command -v wire)" | grep "mod"
```

## Quick Start

For local development, run:

```bash
make dev-up
```

Wait until all services are up and running.

**Services:**
- **API**: http://localhost:8888
- **API Documentation**: http://localhost:8888/docs
- **Localstack**: http://localhost:4566

## Development Commands

Run `make help` to get a list of available commands.

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── app/                # REST API server
│   └── message_processor/  # Event worker/consumer
├── internal/               # Core business logic
│   ├── app/                # Application layer (DI, bootstrap)
│   ├── context/voting/     # Voting bounded context (DDD)
│   ├── integration_test/   # Integration tests
│   └── shared/             # Shared utilities
├── infra/                  # AWS CDK Infrastructure as Code
├── docker/                 # Docker setup for local development
├── config/                 # Configuration management
└── AGENTS.md               # AI Agent reference guide
```

## Contributing

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for details on our code standards, testing requirements, and the process for submitting pull requests.
