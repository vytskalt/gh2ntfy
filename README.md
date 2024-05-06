# gh2ntfy

A lightweight application that forwards GitHub notifications to [ntfy](https://github.com/binwiederhier/ntfy)

## Quickstart

Here's a full docker-compose.yml for gh2ntfy:

```yaml
version: '3'
services:
  gh2ntfy:
    image: ghcr.io/vytskalt/gh2ntfy:0.1.0 # or gh2ntfy-aarch64!
    restart: unless-stopped
    environment:
      GITHUB_TOKEN: 'TODO' # GitHub personal access token with the notifications and repo scopes. Get one at https://github.com/settings/tokens
      NTFY_URL: 'https://ntfy.sh' # Your ntfy server
```

## Building Docker image

Make sure to have the [Nix package manager](https://nixos.org/download/) installed and then run:

```console
nix build .#image # or image-aarch64!
cat result | docker load
```
