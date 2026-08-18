# Vone Decisions

## Task ID

- TASK-20260817-QA-ROCKL-01

## Last Updated

- 2026-08-18 10:27

## Current Decisions

### D1

#### Decision

- Deploy the immutable VONE-0.7.2.1 candidates to the local `vone-weknora` stack in `shadow` after a recoverable PostgreSQL backup. Use an Ubuntu 24.04 compatibility runtime while preserving the exact candidate binary SHA and map its build-time GoJieba dictionary path.

#### Reason

- ROCKL explicitly authorized direct local replacement. The clean Linux candidate requires GLIBC 2.38+ and contains an absolute module-cache dictionary path that the prior Debian 12 image cannot satisfy directly.

#### Impact

- Local app/frontend now run the matching candidates; additive VONE schema version 1 is present and retained during rollback. This is local staging evidence only and does not grant QA, business acceptance or production release approval.

#### Follow-up

- Prepare B48 identities and scoped API key, capture authenticated diagnostics, execute the matrix, and obtain distinct human QA evidence or Tim / T001 explicit scoped exception approval.
