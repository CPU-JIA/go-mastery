# Branch Protection Rules for Go Mastery Project

This document provides recommended branch protection rules to maintain code quality and enforce the CI/CD pipeline requirements.

## GitHub Repository Settings

### Branch Protection Rules for `main` branch

#### 1. Basic Settings
- ✅ **Require a pull request before merging**
  - Required number of approvals: 1
  - ✅ Dismiss stale PR approvals when new commits are pushed
  - ✅ Require review from code owners

#### 2. Status Checks
- ✅ **Require status checks to pass before merging**
- ✅ **Require branches to be up to date before merging**

Required Status Checks:
- `static-analysis (1.21)`
- `static-analysis (1.22)`
- `static-analysis (1.23)`
- `static-analysis (1.24)`
- `security`
- `test (1.21)`
- `test (1.22)`
- `test (1.23)`
- `test (1.24)`
- `integration-tests`
- `build-matrix (linux, amd64)`
- `build-matrix (windows, amd64)`
- `build-matrix (darwin, amd64)`
- `build-matrix (darwin, arm64)`
- `vuln-scan`
- `quality-gate`

#### 3. Additional Restrictions
- ✅ **Require signed commits** (recommended)
- ✅ **Require linear history** (optional)
- ✅ **Include administrators** (enforce rules for admins)
- ✅ **Allow force pushes** ❌ (disabled for safety)
- ✅ **Allow deletions** ❌ (disabled for safety)

#### 4. Push Restrictions
- ✅ **Restrict pushes that create files larger than 100MB**
- ✅ **Block force pushes**

## Repository Rules (GitHub Rulesets)

### Create a Ruleset for Enhanced Protection

```yaml
name: "Go Mastery Quality Gates"
enforcement: active
target: branch
conditions:
  ref_name:
    include:
      - "refs/heads/main"
      - "refs/heads/develop"
rules:
  - type: pull_request
    parameters:
      dismiss_stale_reviews_on_push: true
      require_code_owner_review: true
      required_approving_review_count: 1
      require_last_push_approval: false

  - type: required_status_checks
    parameters:
      strict_required_status_checks_policy: true
      required_status_checks:
        - context: "static-analysis (1.21)"
          integration_id: 15368
        - context: "static-analysis (1.22)"
          integration_id: 15368
        - context: "static-analysis (1.23)"
          integration_id: 15368
        - context: "static-analysis (1.24)"
          integration_id: 15368
        - context: "security"
          integration_id: 15368
        - context: "test (1.21)"
          integration_id: 15368
        - context: "test (1.22)"
          integration_id: 15368
        - context: "test (1.23)"
          integration_id: 15368
        - context: "test (1.24)"
          integration_id: 15368
        - context: "integration-tests"
          integration_id: 15368
        - context: "vuln-scan"
          integration_id: 15368
        - context: "quality-gate"
          integration_id: 15368

  - type: deletion
  - type: non_fast_forward
  - type: commit_message_pattern
    parameters:
      pattern: "^(feat|fix|docs|style|refactor|test|chore|perf|ci)(\\(.+\\))?: .{1,72}"
      flags: []
```

## Environment Protection Rules

### Production Environment
- ✅ **Required reviewers**: 1 admin or designated approver
- ✅ **Wait timer**: 5 minutes (optional cooling-off period)
- ✅ **Deployment branches**: Only `main` branch

## Webhooks Configuration

### Quality Gate Notifications
Set up webhooks for:
- Pull request events
- Push events
- Status events
- Workflow run events

Example webhook payload for quality gate status:
```json
{
  "name": "quality-gate-webhook",
  "config": {
    "url": "https://your-notification-endpoint.com/webhook",
    "content_type": "json",
    "secret": "your-webhook-secret"
  },
  "events": [
    "status",
    "pull_request",
    "workflow_run"
  ]
}
```

## Automated Dependency Updates

### Dependabot Configuration
Create `.github/dependabot.yml`:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    reviewers:
      - "go-mastery-maintainers"
    assignees:
      - "go-mastery-maintainers"
    commit-message:
      prefix: "deps"
      include: "scope"
    open-pull-requests-limit: 5
    rebase-strategy: "auto"
```

## Security Policies

### Code Scanning
- ✅ **CodeQL Analysis**: Enabled for Go
- ✅ **Dependency Review**: Enabled for pull requests
- ✅ **Secret Scanning**: Enabled
- ✅ **Push Protection**: Enabled for secrets

### Security Advisories
- Create security policy in `SECURITY.md`
- Enable private vulnerability reporting
- Configure security advisories

## Repository Settings Checklist

### General
- [ ] Repository visibility: Public/Private as appropriate
- [ ] Features: Issues, Wiki, Projects as needed
- [ ] Merge button: Configure merge options
- [ ] Automatically delete head branches: Enabled

### Access & Security
- [ ] Manage access: Set up teams and permissions
- [ ] Deploy keys: Configure if needed for automation
- [ ] Secrets: Configure CI/CD secrets
- [ ] Actions: Configure repository secrets and variables

### Notifications
- [ ] Email notifications: Configure for maintainers
- [ ] Slack/Teams integration: Set up if desired

## Implementation Commands

### Set Branch Protection via GitHub CLI
```bash
# Install GitHub CLI and authenticate
gh auth login

# Set branch protection for main
gh api repos/:owner/:repo/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["static-analysis","security","test","quality-gate"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
  --field restrictions=null
```

### Verify Protection Settings
```bash
# Check current branch protection
gh api repos/:owner/:repo/branches/main/protection

# List repository rulesets
gh api repos/:owner/:repo/rulesets
```

## Quality Gate Success Criteria

For a pull request to be mergeable, it must pass:

1. ✅ All required status checks
2. ✅ At least 1 approving review
3. ✅ No unresolved conversations
4. ✅ Branch is up to date with main
5. ✅ All CI/CD pipeline stages complete successfully
6. ✅ Coverage threshold maintained (≥75%)
7. ✅ Security scans pass with no HIGH/MEDIUM vulnerabilities
8. ✅ Code formatting and linting checks pass

## Monitoring and Alerting

### Key Metrics to Monitor
- Build success rate
- Test coverage percentage
- Security vulnerability count
- Code quality metrics
- Deployment frequency
- Lead time for changes
- Mean time to recovery

### Recommended Tools
- GitHub Insights for repository analytics
- Codecov for coverage tracking
- Snyk for security monitoring
- SonarCloud for code quality metrics

---

**Note**: Adjust these settings based on your team's specific needs and organizational policies. The goal is to maintain high code quality while enabling efficient development workflows.