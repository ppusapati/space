# DPIA index

> **TASK-P0-COMP-001 (PR-G).**
> Index of every Data Protection Impact Assessment (GDPR Art. 35).
> Each row links to the specific DPIA document; the template lives at
> [`template.md`](template.md). DPIAs are filed by the responsible
> service team and signed off by the DPO before the processing goes
> live (see TASK-P5-COMP-001 + TASK-P6-COMP-002 in `plan/todo.md`).

| DPIA ID | Scope | Owner | Linked task | Status | File |
|---|---|---|---|---|---|
| DPIA-0001 | Public Imagery-as-a-Service customer surface (sign-up, AOI subscription, deliveries, downloads, DOI registration) | Compliance + EO | [TASK-P5-COMP-001](../../plan/todo.md) | blocked:OQ-009 | `dpia-iaas.md` (to be filed in PR-G consumer flow) |
| DPIA-0002 | Platform-wide processing (IAM, audit, notify, export, realtime-gw) | Compliance + Platform IAM | [TASK-P6-COMP-002](../../plan/todo.md) | blocked:OQ-009 | `dpia-platform.md` (to be filed in Phase 6) |

## How to file a new DPIA

1. Copy `template.md` to `dpia-<scope>.md`.
2. Fill in §1 identification with a fresh DPIA-ID (next free number above).
3. Walk through §§2–6 with the responsible service owner.
4. Submit for DPO review (`compliance/sign-offs/<dpia-id>.pdf` once signed).
5. Add the row to the table above.
6. Reference the DPIA-ID from the relevant ROPA entry in
   [`../ropa.md`](../ropa.md).
