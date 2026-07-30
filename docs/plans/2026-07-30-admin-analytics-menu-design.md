# Admin Analytics Menu Design

## Goal

Make the existing admin analytics page directly discoverable from the wallet dropdown.

## Design

Keep the existing **Admin** moderation item and pending-review badge unchanged. Add a separate **Analytics** item immediately below it, visible under the same `isAdmin` condition, linking to `/admin/analytics` and closing the dropdown when selected.

## Verification

Add a source-contract test for the admin-only route, then run the frontend test suite and production build.
