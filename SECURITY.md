# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 0.1.x   | Yes       |

## Reporting a Vulnerability

Please report security vulnerabilities privately by emailing
**sovletig@gmail.com** with:

- a description of the vulnerability and its impact,
- steps to reproduce or a proof of concept,
- affected versions, if known.

You can expect an acknowledgement within 72 hours and a status update
within 7 days. If the report is confirmed, a fix will be released as
soon as practical and credited (if desired) in the release notes.

Please do not open public GitHub issues for security reports.

## Scope

The SDK itself never handles credentials beyond the caller-supplied
CoinGecko Pro API key, which is sent only in the `x-cg-pro-api-key`
header and never placed in URLs, logs or error payloads.
