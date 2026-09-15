# Presence example

The presence example connects to a Galene group and reports when users start or stop streaming.

It prints sharing-specific messages for `camera` and `screenshare` streams. Streams without a label, including WHIP streams, are reported as generic streaming activity.

## Usage

```text
go run ./examples/presence [options] group
```

For example:

```text
go run ./examples/presence \
  -username alice \
  -password secret \
  https://galene.example.org/group/example/
```

Options:

- `-username`: username used to join the group; defaults to `presence-example`.
- `-password`: password used to join the group.
- `-insecure`: disable TLS certificate verification.
- `-debug`: enable Galene protocol logging.

The example requests audio and video streams so Galene sends stream offers. It does not read or process media packets; it uses the down-connection and close events to report presence.
