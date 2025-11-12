# Contributing to Recs Votes Storage

Thank you for contributing to the Recs Votes Storage service! This guide will help you understand our development workflow and standards.

## Code of Conduct

- Be respectful and professional in all interactions
- Follow Bumble's engineering principles and best practices
- Ask questions when unclear - no question is too small

## Development Workflow

### 1. Branch Naming

Follow the pattern: `RECS-{ticket-number}` or `RECS-{ticket-number}-brief-description`

Example:
```bash
git checkout -b RECS-1234
git checkout -b RECS-1234-add-vote-validation
```

### 2. Making Changes

Before starting development, ensure your environment is set up:

```bash
# Install dependencies (see README.md for details)
make wire            # Regenerate DI if needed
make fmt             # Format your code
make lint            # Check for linting issues
make test            # Ensure all tests pass
```

### 3. Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) format with the ticket number at the end:

```
<type>(<scope>): <description> [RECS-XXXX]

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `docs`: Documentation changes
- `chore`: Maintenance tasks (dependency updates, etc.)
- `perf`: Performance improvements

**Examples:**
```
feat(votes): add vote type validation for crush votes [RECS-1234]

fix(counters): resolve race condition in hourly counter updates [RECS-1235]

refactor: change constructors to return pointers for consistency [RECS-1473]

docs(readme): add installation prerequisites [RECS-1475]
```

### 4. Pull Requests

Create a pull request against the `master` branch with:
- Clear title following conventional commit format
- Description of what changed and why
- Link to relevant Jira ticket (RECS-XXXX)
- Screenshots/recordings if UI changes

## Code Requirements

### Testing (Required)

**All new code must include tests.** We maintain a 95% code coverage target.

- **Unit Tests**: For business logic, operations, and utilities
- **Integration Tests**: For repository implementations, database operations, AWS service interactions
- **Table Tests**: Use Go's table-driven test pattern where appropriate

```go
// Example table-driven test
func TestVoteTypeValidation(t *testing.T) {
    tests := []struct {
        name    string
        voteType int
        want    bool
        wantErr bool
    }{
        {"valid yes vote", 1, true, false},
        {"invalid vote type", 99, false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Run tests before submitting:**
```bash
make test              # Run all tests
make test-coverage     # Check coverage meets 95% threshold
```

### Documentation Requirements

Update `AGENTS.md` when your changes include:

- ✅ **New API endpoints** - Add to API Endpoints section
- ✅ **New message types or handlers** - Update Event-Driven Architecture section
- ✅ **New operations** - Add to Application Operations section
- ✅ **Database schema changes** - Update Database Schema section
- ✅ **New coding patterns or conventions** - Add to Coding Conventions section
- ✅ **Architecture changes** - Update relevant sections
- ✅ **New features** - Update appropriate sections with examples

**Minor changes that don't require AGENTS.md updates:**
- Bug fixes that don't change behavior
- Refactoring without interface changes
- Test additions/improvements
- Code formatting or linting fixes

#### Example Command

Tell Claude something like:
- `"Update AGENTS.md"`
- `"Update AGENTS.md to reflect recent code changes"`

### Code Style and Standards

#### 1. Go Standard Practices

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (enforced by `make fmt`)
- Run `golangci-lint` (enforced by `make lint`)
- Keep functions small and focused (< 50 lines when possible)
- Use meaningful variable names (avoid single-letter except in loops)

#### 2. Project-Specific Conventions

**Constructor Functions and Return Types:**

Follow Go best practices for constructor return types based on the struct characteristics:

**Return Pointers When:**
- The struct has methods with pointer receivers
- The struct is mutable or holds state
- The struct is large (copying would be expensive)
- The struct will be used to satisfy interfaces with pointer receiver methods

**Return Values When:**
- The struct is small and immutable (value objects, simple DTOs)
- The struct has no methods or only value receiver methods
- The struct represents a simple data container

```go
// ✅ Correct - Service with pointer receivers
func NewVotingService(...) *VotingService {
    return &VotingService{...}
}

// ✅ Correct - Small immutable value object
func NewVoteId(id uuid.UUID) VoteId {
    return VoteId{value: id}
}

// ❌ Incorrect - Service with pointer receivers returning value
func NewVotingService(...) VotingService {
    return VotingService{...}
}
```

**Pointer Receivers for Methods:**
```go
// ✅ Correct - consistent pointer receivers
func (s *VotingService) AddVote(...) error { }
func (s *VotingService) DeleteVote(...) error { }

// ❌ Incorrect - mixing value and pointer receivers
func (s VotingService) AddVote(...) error { }
func (s *VotingService) DeleteVote(...) error { }
```

**Error Handling:**

Wrap errors with context when it adds real value, typically at:
- **Subsystem boundaries** (API → service → repository)
- **Places where the root cause isn't obvious** from the call stack
- **When adding business context** that explains what operation failed

Don't wrap errors when:
- The error message is already clear and self-explanatory
- You're just passing it up one level in the same layer
- Wrapping would be redundant

```go
// ✅ Correct - wrapping at subsystem boundary adds context
func (s *VotingService) AddVote(ctx context.Context, vote Vote) error {
    romance, err := s.repo.GetRomance(ctx, vote.ActiveUserKey, vote.PeerUserKey)
    if err != nil {
        return fmt.Errorf("failed to get romance for vote addition: %w", err)
    }
    // ...
}

// ✅ Correct - error is already clear, no wrapping needed
func (r *DynamoDBRepository) GetRomance(ctx context.Context, key1, key2 UserKey) (*Romance, error) {
    result, err := r.client.GetItem(ctx, input)
    if err != nil {
        return nil, err  // DynamoDB error is already descriptive
    }
    // ...
}

// ❌ Incorrect - redundant wrapping that doesn't add value
func helper(repo Repository) error {
    err := repo.Save(item)
    if err != nil {
        return fmt.Errorf("failed to save: %w", err)  // "failed to save" adds no context
    }
    return nil
}
```

**Dependency Injection:**
- Use Google Wire for dependency injection
- All dependencies injected via constructors
- Run `make wire` after changing DI configuration

```go
// Constructor with injected dependencies
func NewVotingService(
    repo RomancesRepository,
    logger Logger,
) *VotingService {
    return &VotingService{
        repo:   repo,
        logger: logger,
    }
}
```

#### 3. Project Structure

Follow Clean Architecture and DDD principles:

```
internal/context/voting/
├── domain/           # Business entities and rules (no external dependencies)
├── application/      # Use cases and business logic orchestration
├── infrastructure/   # External concerns (database, messaging)
└── interface/        # API contracts and handlers
```

**Dependency Rule:** Code dependencies must point inward
- Domain has no external dependencies
- Application depends on domain
- Infrastructure and interface depend on application and domain

#### 4. Naming Conventions

- **Interfaces**: Suffix with behavior (`Repository`, `Publisher`, `Logger`)
- **Implementations**: Prefix with tech (`DynamoDBRepository`, `SnsPublisher`)
- **Operations**: Suffix with `Operation` (`AddUserVoteOperation`)
- **Handlers**: Suffix with `Handler` (`DeleteRomancesHandler`)
- **Messages**: Suffix with `Message` (`DeleteRomancesMessage`)

#### 5. Code Organization

**Order of declarations in a file:**

1. Package declaration
2. Imports (stdlib, external, internal)
3. Constants
4. Type definitions
5. Constructor functions
6. Methods (grouped by receiver)
7. Private helper functions

## Getting Help

- **Jira Issues**: [RECS Project Board](https://bmbl.atlassian.net/browse/RECS)
- **Design Docs**: See README.md for links to Confluence and Figma

## Review Process

1. **Automated Checks**: CI will run tests, linting, and coverage checks
2. **Peer Review**: At least one approval required from team member
3. **Code Owner Review**: May be required for certain paths
4. **Merge**: Squash and merge to keep git history clean

## Common Mistakes to Avoid

❌ **Don't:**
- Push directly to `master`
- Commit without running tests
- Use `// nolint` directives without good reason
- Mix refactoring with feature changes in the same commit
- Return values from constructors that have pointer receiver methods
- Skip error handling or return generic errors
- Mutate shared state without synchronization

✅ **Do:**
- Write tests first (TDD when applicable)
- Keep PRs focused and reasonably sized
- Add context in commit messages
- Return pointers from constructors for consistency
- Wrap errors with meaningful context
- Use channels for streaming large datasets
- Document complex business logic
