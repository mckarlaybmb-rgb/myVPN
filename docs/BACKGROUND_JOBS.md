# Background Jobs

The PostgreSQL `job_queue` is consumed by the in-process worker. Jobs are claimed with `FOR UPDATE SKIP LOCKED`, retried up to three attempts, and then marked failed. Handlers must be idempotent; unsupported job types are terminal failures.

The expiry scheduler runs hourly and processes subscriptions expiring within seven days. It disables expired VPN clients, updates subscription state, and sends deduplicated Telegram notifications. The node checker runs every five minutes, probes active nodes with TCP dial timeouts, records latency and status, and isolates node failures from the API process.

For production, run one API process for the scheduler or add a distributed lease before running multiple replicas. Monitor failed jobs, unhealthy nodes, scheduler logs, and Telegram delivery failures.
