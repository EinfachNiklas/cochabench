# Security Policy

## Supported Versions

CochaBench is currently under active development. Security fixes are applied to the latest version on the default branch.

## Reporting a Vulnerability

Please do not report security issues in public GitHub issues.

Report vulnerabilities privately to the maintainer through a private channel if one is available. If no private reporting channel is configured yet, open a minimal public issue without exploit details and request a private contact path.

When reporting an issue, include:

- A short description of the vulnerability
- Affected version or commit
- Reproduction steps
- Expected impact
- Any suggested mitigation

## Operational Risk Notice

CochaBench executes challenge code, test suites, and package installation commands on the local machine.

This means:

- Untrusted challenges can execute arbitrary code with the permissions of the current user
- JavaScript evaluation may run `npm install` and `npm test`
- Python evaluation may create virtual environments and install packages with `pip`
- Go evaluation may download modules and execute project tests

CochaBench does not currently provide built-in sandboxing or container isolation. For untrusted inputs, use an isolated environment such as a disposable VM or container.

## Disclosure Process

After receiving a valid report, the goal is to:

- Confirm the issue
- Assess severity and affected scope
- Prepare and validate a fix
- Publish the fix and disclose the issue responsibly
