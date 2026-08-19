# Job Queue

Status: Durable queue primitives implemented; worker planned

The durable queue is stored in PostgreSQL table `job_queue`.

## States

- `pending`: available for a worker to claim.
- `processing`: claimed by a worker and currently being handled.
- `completed`: handled successfully.
- `failed`: handling ended unsuccessfully and the error is recorded.

## Processing Semantics

`internal/jobs.Queue.Claim` selects an available pending row using `FOR UPDATE SKIP LOCKED`, changes it to `processing`, and increments `attempts` in one transaction. `Complete` and `Fail` transition processing jobs to terminal states.

Phase 1 includes the durable queue operations and subscription enqueue path. No long-running worker consumes these rows yet.

## Planned

Retry backoff, dead-letter workflow, stale-processing recovery, queue metrics, and a long-running worker remain follow-up work.
