# Published container images

Requested by a user (2026-10-29): a published Docker image.

## Summary

GitHub Actions builds and pushes to the GitHub Container Registry:
`ghcr.io/kallisti-chat/kallisti` and `ghcr.io/kallisti-chat/kallisti-cert-sync`,
for linux/amd64 and linux/arm64. Version tags (v0.10.1) give `0.10.1`,
`0.10` and `latest`; each push to main gives `edge`. The web and Go
stages build on the runner's platform (Go cross-compiles); only the last
stage runs emulated. A compose override uses the published image (our
own servers keep building from source); the Helm chart gets the images
as defaults; the install guide says how to pull.

After the first push the packages are private (GitHub's default for an
organisation): the owner makes them public once in the package settings.

## Acceptance Criteria

- The Dockerfile still builds locally; the workflow is valid (actionlint).
- Compose with the override pulls the image; the chart test passes.

## Result (2026-10-29)

Built: workflow image.yml (docker/metadata-action: semver tags and latest on v*, edge on main), the Dockerfile cross-builds, compose.published.yml swaps Kallisti for the image, cert-sync has image and build in compose.yml (pull or build). The Helm chart keeps its rule: a fixed version, no latest; its README names the images. Checked: podman build arm64 and the amd64 server stage (an x86-64 binary), compose config with each compose file, actionlint. Open: the first run on GitHub, then the packages made public.
