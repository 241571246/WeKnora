# Vone Decisions

## Task ID

- TASK-20260817-SEC-ROCKL-01

## Last Updated

- 2026-08-17 20:48

## Current Decisions

### D1

#### Decision

- Retain existing `kb_shares` cross-workspace semantics, but route every shared access decision through the central KB Authorizer. Do not convert organization shares into employee memberships.

#### Reason

- Organization sharing and same-workspace employee assignment are different business relationships; conversion would lose semantics, while the current direct path would bypass fine-grained ACL.

#### Impact

- Backend policy must combine workspace role, KB membership capability, organization share, Agent context and API-key scope under default deny.

#### Follow-up

- Implement and verify in B3, B13, B16 and B28.

### D2

#### Decision

- Treat `ROCKL` and `RockL` as the same human identity. Record RockL as QA Owner but set QA Independence Gate to `EXCEPTION REQUIRED`.

#### Reason

- Case differences do not create process independence, and REQ-2026-001 is High risk.

#### Impact

- Technical work may proceed after admission, but QA Passed and Release Authorization require a distinct human reviewer or a scoped exception approved by another named human.

#### Follow-up

- Keep ACT-2026-001 open until the independent reviewer or exception approver is recorded.
