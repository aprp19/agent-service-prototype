# Agent Service — Database Migration Design

## Background

We operate applications that are deployed across multiple on-premise environments. Each installation can have its own database server, which makes schema updates difficult to coordinate consistently.

Without centralized orchestration, database updates may require engineers to connect to individual environments and execute SQL scripts manually. This approach introduces several risks:

- inconsistent schema versions between environments
- migrations being executed more than once
- partially completed database changes
- limited visibility into migration history
- higher operational effort as the number of installations grows

To explore a safer and more scalable approach, I built a proof of concept where a centralized command center coordinates lightweight agents running inside each on-premise environment.

The command center is responsible for orchestration, while the agent performs the actual database migration locally.

## Goals

The prototype was designed around several goals:

- make database migrations repeatable and traceable
- avoid exposing on-premise databases directly to a central service
- prevent multiple migration processes from modifying the same database concurrently
- detect unexpected changes to migration files
- support both fresh installations and incremental upgrades
- provide enough execution state for operators to diagnose failures
- validate the architecture before building a production-ready implementation

## Proposed Architecture

The system is conceptually divided into two components.

### Command Center

The command center is responsible for coordinating database deployments across multiple environments.

Its responsibilities include:

- determining which database version should be deployed
- sending installation commands to the appropriate agent
- tracking deployment status
- collecting execution results from agents

The command center does not need direct access to the target database.

### On-Premise Agent

The agent runs inside the target environment and performs database operations locally.

Its responsibilities include:

- downloading the requested database bundle
- verifying bundle integrity
- inspecting the current database state
- applying baseline and incremental migrations
- preventing concurrent migration execution
- recording migration history
- performing post-migration checks
- reporting execution status

This approach keeps database credentials and connectivity inside the local environment while still allowing deployment orchestration to be centralized.

## Migration Workflow

When the agent receives an installation command, it performs the following steps:

1. Download the versioned database bundle.
2. Extract the bundle into a temporary working directory.
3. Verify file integrity using SHA-256 checksums.
4. Parse the migration manifest.
5. Connect to the local PostgreSQL database.
6. Acquire a PostgreSQL advisory lock.
7. Detect whether the database is fresh or already initialized.
8. Apply the baseline schema when necessary.
9. Apply pending incremental migrations.
10. Record migration results and execution time.
11. Run post-migration smoke checks.
12. Release the database lock.

Each step is tracked independently so that failures can indicate exactly where the installation stopped.

## Database Bundle

A database release is distributed as a versioned bundle.

The bundle can contain:

- a baseline schema
- incremental migration files
- a manifest describing migration order and metadata
- checksum metadata
- optional smoke-check SQL

The manifest allows the agent to execute migrations in a deterministic order instead of relying on file-system ordering or manually selected scripts.

## Handling Fresh and Existing Databases

The agent must support two different database states.

### Fresh Database

If the database has not been initialized, the agent applies the baseline schema first.

The baseline represents the starting schema for a new installation.

After the baseline is applied, the agent creates or verifies the migration history table and then continues with any required incremental migrations.

### Existing Database

If the database is already initialized, the agent skips the baseline and checks the migration history to determine which migrations still need to be executed.

This makes the same installation workflow usable for both new environments and existing deployments.

## Preventing Concurrent Migrations

One of the problems I wanted to avoid was having two migration processes modify the same database simultaneously.

The agent therefore acquires a PostgreSQL advisory lock before applying database changes.

Using a database-level lock provides an additional safety mechanism even if multiple installation commands are accidentally triggered at the same time or more than one agent process attempts to perform an installation.

Only one migration process can hold the lock for the configured key at a time.

The lock is released when the installation process finishes.

## Migration Integrity

Migration files are treated as immutable after they have been applied.

Before executing a migration, the agent calculates the SHA-256 checksum of the migration file.

The checksum is stored together with the migration history.

If the same migration version already exists in the database but the current file has a different checksum, the agent stops the installation instead of silently executing the modified script.

This protects against a common source of schema drift: changing historical migration files after they have already been executed in another environment.

## Idempotency

The installation process is designed to be safe to retry.

Before executing a migration, the agent checks whether that migration version has already been successfully applied.

If it has already completed successfully, the migration is skipped.

This allows an installation command to be retried without automatically re-running all previously completed database changes.

A force option can still be provided for controlled scenarios where a migration must intentionally be reprocessed.

## Transaction Handling

Not every PostgreSQL operation can or should run inside a transaction.

For that reason, migrations explicitly declare whether they should be transactional.

### Transactional Migration

The SQL is executed inside a database transaction.

If execution fails, the transaction can be rolled back.

This is the preferred approach for migrations where all schema changes should succeed or fail as a single unit.

### Non-Transactional Migration

Some database operations may need to run outside a transaction.

For these cases, the migration is executed directly while its execution result is still recorded in the migration history.

Making transaction behavior explicit avoids assuming that every database operation has identical transactional requirements.

## Migration History

For each migration, the agent records metadata such as:

- migration version
- migration name
- checksum
- execution timestamp
- execution duration
- success or failure state
- error message when applicable

This provides an audit trail that can be used to understand the current schema state and investigate failed deployments.

It also allows the agent to determine which migrations are still pending during future installations.

## Failure Handling

The installation process tracks its current execution step.

Examples include:

- `DOWNLOAD_BUNDLE`
- `EXTRACT_BUNDLE`
- `VERIFY_CHECKSUM`
- `PARSE_MANIFEST`
- `CONNECT_DB`
- `LOCK_DB`
- `APPLY_BASELINE`
- `APPLY_MIGRATIONS`
- `POST_CHECK`

If the process fails, the agent reports both the failed step and the associated error.

This is useful for the command center because it can distinguish between infrastructure failures, invalid bundles, database connectivity problems, migration errors, and failed validation checks.

## Post-Migration Validation

After the migration sequence completes, the agent can execute a smoke-check SQL file.

The purpose of the smoke check is not to perform a complete application test, but to verify basic assumptions about the resulting database state.

For example, a smoke check could verify that an expected table, column, or schema object exists after the migration.

If the smoke check fails, the overall installation is reported as failed.

## Alternatives Considered

### Executing SQL Directly from the Command Center

One possible approach would be for the command center to connect directly to every database server.

I avoided this design because it would require exposing database connectivity across environments and centrally managing database credentials.

Running an agent locally keeps database access inside the on-premise environment and reduces the amount of infrastructure that needs direct database connectivity.

### Running Migrations Manually

Manual migration execution is simple for a small number of environments but becomes difficult to manage as deployments increase.

It also makes it harder to guarantee that every environment receives exactly the same migration sequence.

The agent-based approach introduces additional system complexity, but provides consistency, automation, and better observability.

### Relying Only on Migration Version Numbers

Another option would be to track only migration version numbers.

I chose to also track checksums because version numbers alone cannot detect whether an already-applied migration file has been modified.

Checksum validation provides an additional integrity guarantee.

## Trade-Offs

The agent-based architecture introduces some additional complexity.

The system now needs to handle:

- communication between the command center and agents
- agent lifecycle management
- database bundle distribution
- status synchronization
- compatibility between agent versions and database bundles
- authentication and authorization between components

However, this complexity is intentional.

The goal is to move operational complexity away from manual procedures and into a system that can enforce consistent behavior across many environments.

## Future Improvements

The current implementation is a proof of concept intended to validate the overall migration model.

Possible next steps include:

- signed database bundles
- mutual TLS between the command center and agents
- stronger command authentication
- deployment retry policies
- richer rollback strategies
- structured event reporting
- centralized migration dashboards
- agent version compatibility checks
- resumable bundle downloads
- distributed deployment scheduling
- deployment approval workflows
- better observability with metrics and tracing

## Why I Built This Prototype

What made this project interesting to me was that I was not simply implementing an existing specification.

I first had to research the operational problem and evaluate different approaches for safely managing database deployments across distributed on-premise environments.

That required thinking about:

- migration versioning
- concurrency
- idempotency
- integrity verification
- transaction boundaries
- failure recovery
- observability
- local infrastructure constraints
- centralized orchestration without direct database access

The result is a working prototype that helps validate the architecture before the company invests in a larger production implementation.

This project is one of the pieces of work I am most proud of because it required both implementation and system-level engineering decisions rather than only feature development.

## Reference Implementation

Repository:

https://github.com/aprp19/agent-service-prototype
