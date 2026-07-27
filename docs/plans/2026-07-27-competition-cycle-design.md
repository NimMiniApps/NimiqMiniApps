# Competition Cycle Design

## Goal

Record which Nimiq Mini Apps competition cycle an app belongs to while keeping
the catalog open to apps that are not competition entries. Show competition
membership publicly and let visitors filter the catalog by cycle.

## Data model and API

Add `competition_cycle` to apps and app revisions as a nullable integer. A null
value means the app is not a competition entry; positive integers identify
cycles such as Cycle 1. Reject zero and negative values.

Expose the field in public and admin app responses, developer submissions,
update requests, and admin create/update operations. Add
`competition_cycle=<positive integer>` to `GET /api/apps` for server-side
filtering and accept `competition_cycle=none` for non-competition apps. Keep the
API open to future positive cycle numbers so new cycles do not require a schema
migration.

Backfill Cycle 1 for catalog records that match the verified Cycle 1 submission
repository. Entries that are not yet in the catalog can select Cycle 1 when
they are submitted later.

Update the OpenAPI source and regenerate both backend copies. Add the field to
the MCP create/update tools and document its meaning.

## User interface

Submission, update-request, and admin forms include an optional Competition
cycle control. Leaving it empty means the app is not a competition entry.

App cards and app detail pages show a `Cycle N` badge when the value is set. The
Apps page includes a cycle filter alongside its existing category, developer,
and sort controls. The filter offers all competition cycles represented in the
loaded catalog and an option for apps that are not competition entries.

## Data flow

The backend validates and stores the selected cycle during create and submit.
Update requests copy the value into app revisions, and revision approval copies
it back to the app. API consumers normalize a missing field to `null` for
compatibility during deployment.

## Validation and errors

Only positive integers are valid cycle values. Invalid values return HTTP 400
through the existing app validation path. An omitted or null value is valid.
The list endpoint rejects malformed or non-positive cycle query values rather
than silently returning an unfiltered catalog.

## Testing

Backend tests cover accepted nullable and positive values, rejected non-positive
values, cycle query parsing, SQL filtering, and revision field preservation.
Frontend tests cover compatibility normalization, badge output, and cycle
filter state/query behavior. Final verification runs backend tests and build,
MCP tests/build, frontend tests/build, and the OpenAPI synchronization check.

## Documentation

Update README, AGENTS, and MCP documentation because competition cycle selection
is visible to developers, catalog visitors, moderators, and MCP clients.
