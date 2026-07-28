# Video-First Fullscreen Media Design

## Goal

Make videos the first items users encounter in an app's media gallery and let users inspect submitted screenshots at their full available resolution.

## Interaction

- Display all YouTube media before image media, preserving the submitted order within each media type.
- Keep thumbnail clicks dedicated to selecting the featured media item.
- Make the featured image an accessible button that opens a native modal dialog.
- Show the original image URL inside the dialog with `object-contain`, a dark backdrop, a visible close button, backdrop-click dismissal, and native Escape-key dismissal.
- Keep YouTube playback and fullscreen behavior inside the existing YouTube embed.

## Architecture

Add a pure stable-ordering helper to the existing media-gallery utility module and use its result as the gallery component's display list. Keep the original prop immutable. The gallery owns a dialog element reference and opens or closes it through the native `<dialog>` API.

## Testing

- Unit-test that videos move before images without changing order inside either group.
- Assert that the component renders from the ordered list.
- Assert that the featured image is a labelled button connected to a native dialog containing the full-resolution image and close control.
- Run all frontend tests and the production build before release.

## Deployment

Commit and push to `main`, wait for GitHub Actions to publish the exact SHA-tagged frontend image, update only `nimiqminiapps_frontend`, and verify the Swarm service, public health endpoints, app-detail route, and live bundle.
