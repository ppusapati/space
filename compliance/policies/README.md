# ISMS policies

> **TASK-P0-COMP-001 (PR-G).**
> Index of every ISMS policy referenced by the control register at
> [`../controls/iso27001.csv`](../controls/iso27001.csv) and the
> staged-certification CSVs. Files themselves are authored in
> **TASK-P6-COMP-001** (Phase 6 — ISMS skeleton finalised). Each row
> below is the canonical filename + the ISO 27001 controls it backs.

## How this register works

- One Markdown file per policy. Filename is the slug used in the
  `evidence_path` column of every control CSV — so all references stay
  resolveable as policies land.
- Every policy follows the structure: **Purpose · Scope · Roles ·
  Procedures · Compliance · Review.**
- Reviewed annually as part of the management-review cycle (ISO 27001
  A.5.4).
- Policy changes require Compliance + Engineering owner sign-off via
  PR review.

## Policy register

| Policy file | ISMS family | Backs controls | Owner | Status |
|---|---|---|---|---|
| `information-security.md`   | A.5 organisational  | A.5.1                              | Compliance       | pending (P6) |
| `roles.md`                  | A.5 organisational  | A.5.2                              | Compliance       | pending (P6) |
| `segregation.md`            | A.5 organisational  | A.5.3                              | Compliance       | pending (P6) |
| `management-review.md`      | A.5 organisational  | A.5.4                              | Compliance       | pending (P6) |
| `authorities.md`            | A.5 organisational  | A.5.5                              | Compliance       | pending (P6) |
| `sigs.md`                   | A.5 organisational  | A.5.6                              | Compliance       | pending (P6) |
| `threat-intel.md`           | A.5 organisational  | A.5.7                              | Security         | pending (P6) |
| `project-security.md`       | A.5 organisational  | A.5.8                              | Platform Arch    | pending (P6) |
| `asset-inventory.md`        | A.5 organisational  | A.5.9, A.7.13                      | Platform Infra   | pending (P6) |
| `aup.md`                    | A.5 organisational  | A.5.10                             | Platform Infra   | pending (P6) |
| `asset-return.md`           | A.5 organisational  | A.5.11                             | Platform Infra   | pending (P6) |
| `information-transfer.md`   | A.5 organisational  | A.5.14                             | Security         | pending (P6) |
| `supplier.md`               | A.5 organisational  | A.5.19, A.5.20, A.5.22             | Compliance       | pending (P6) |
| `legal.md`                  | A.5 organisational  | A.5.31                             | Compliance       | pending (P6) |
| `ip.md`                     | A.5 organisational  | A.5.32                             | Compliance       | pending (P6) |
| `screening.md`              | A.6 people          | A.6.1                              | Compliance       | pending (P6) |
| `employment.md`             | A.6 people          | A.6.2                              | Compliance       | pending (P6) |
| `awareness-training.md`     | A.6 people          | A.6.3                              | Compliance       | pending (P6) |
| `disciplinary.md`           | A.6 people          | A.6.4                              | Compliance       | pending (P6) |
| `offboarding.md`            | A.6 people          | A.6.5                              | Compliance       | pending (P6) |
| `nda.md`                    | A.6 people          | A.6.6                              | Compliance       | pending (P6) |
| `remote-work.md`            | A.6 people          | A.6.7                              | Compliance       | pending (P6) |
| `physical-security.md`      | A.7 physical        | A.7.1–A.7.5, A.7.8, A.7.11, A.7.12 | Platform Infra   | pending (P6) |
| `clear-desk.md`             | A.7 physical        | A.7.7                              | Compliance       | pending (P6) |
| `byod.md`                   | A.7 physical        | A.7.9, A.8.1                       | Compliance       | pending (P6) |
| `disposal.md`               | A.7 physical        | A.7.14                             | Platform Infra   | pending (P6) |
| `change-management.md`      | A.8 technological   | A.8.32                             | Platform Infra   | pending (P6) |
| `cryptography.md`           | A.8 technological   | A.8.24                             | Security         | pending (P6) |
| `incident-response.md`      | A.5 organisational  | A.5.24–A.5.27, A.6.8               | Security         | pending (P6) |
| `bcp.md`                    | A.5 organisational  | A.5.29, A.5.30                     | Platform Infra   | pending (P6) |
| `secure-development.md`     | A.8 technological   | A.8.25, A.8.28                     | Security         | pending (P6) |
| `vulnerability-management.md` | A.8 technological | A.8.8                              | Security         | pending (P6) |
| `risk-management.md`        | A.5 organisational  | (cross-cutting)                    | Compliance       | pending (P6) |
| `information-classification.md` | A.5 organisational | A.5.12, A.5.13                  | Platform Arch    | implemented (`compliance/classification.yaml`) |
| `network-services.md`       | A.8 technological   | A.8.21                             | Platform Infra   | pending (P6) |
| `web-filtering.md`          | A.8 technological   | A.8.23                             | Compliance       | pending (P6) |
| `outsourced-development.md` | A.8 technological   | A.8.30                             | Compliance       | pending (P6) |
| `test-data.md`              | A.8 technological   | A.8.33                             | Platform Infra   | pending (P6) |
| `audit-testing.md`          | A.8 technological   | A.8.34                             | Security         | pending (P6) |
| `privileged-utilities.md`   | A.8 technological   | A.8.18                             | Compliance       | pending (P6) |
| `ntp.md`                    | A.8 technological   | A.8.17                             | Platform Infra   | pending (P6) |
| `k8s-admission.md`          | A.8 technological   | (consumes A.8.19 + image signing)   | Platform Infra   | pending (P6) |
| `dns.md`                    | A.8 technological   | (FedRAMP SC-20/21/22)              | Platform Infra   | pending (P6) |
| `email.md`                  | A.8 technological   | (FedRAMP SI-8)                     | Compliance       | pending (P6) |
| `mobile-code.md`            | A.8 technological   | (FedRAMP SC-18)                    | Compliance       | pending (P6) |
| `software-usage.md`         | A.8 technological   | (FedRAMP CM-10/11)                 | Compliance       | pending (P6) |
| `media-protection.md`       | A.7 physical        | (FedRAMP MP family)                | Compliance       | pending (P6) |
| `personnel-security.md`     | A.6 people          | (FedRAMP PS family)                | Compliance       | pending (P6) |
| `planning.md`               | A.5 organisational  | (FedRAMP PL family)                | Compliance       | pending (P6) |
| `privacy.md`                | A.5 organisational  | (FedRAMP PT family)                | Compliance       | pending (P6) |
| `system-integrity.md`       | A.8 technological   | (FedRAMP SI family)                | Security         | pending (P6) |
| `supply-chain.md`           | A.5 organisational  | (FedRAMP SR family)                | Security         | pending (P6) |
| `assessment.md`             | A.5 organisational  | (FedRAMP CA family)                | Compliance       | pending (P6) |
| `maintenance.md`            | A.8 technological   | (FedRAMP MA family)                | Platform Infra   | pending (P6) |
| `external-systems.md`       | A.5 organisational  | (FedRAMP AC-20)                    | Compliance       | pending (P6) |
| `information-sharing.md`    | A.5 organisational  | (FedRAMP AC-21, CA-3)              | Compliance       | pending (P6) |
| `vuln-disclosure.md`        | A.5 organisational  | (FedRAMP RA-5(11))                 | Security         | pending (P6) |
| `wireless.md`               | A.8 technological   | (FedRAMP AC-18)                    | Compliance       | pending (P6) |
| `code-of-conduct.md`        | A.6 people          | (SOC 2 CC1.1)                      | Compliance       | pending (P6) |
| `communications.md`         | A.5 organisational  | (SOC 2 CC2)                        | Compliance       | pending (P6) |
| `identity-proofing.md`      | A.6 people          | (FedRAMP IA-12)                    | Compliance       | pending (P6) |
| `audit-policy.md`           | A.8 technological   | (FedRAMP AU-1)                     | Security         | pending (P6) |
| `access-control.md`         | A.5 organisational  | (cross-cutting AC + IA)            | Platform IAM     | pending (P6) |
| `itar-agreements.md`        | (ITAR)              | 22 CFR 124                         | Compliance       | pending (v2.0) |
| `itar-technical-data.md`    | (ITAR)              | 22 CFR 125                         | Compliance       | pending (v2.0) |
| `itar-general.md`           | (ITAR)              | 22 CFR 126                         | Compliance       | pending (v2.0) |
| `itar-violations.md`        | (ITAR)              | 22 CFR 127                         | Compliance       | pending (v2.0) |
| `source-code-access.md`     | A.5 organisational  | A.8.4                              | Platform Infra   | pending (P6) |
