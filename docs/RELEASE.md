# Release and Registry Publishing

This provider is published as `dnscaleou/dnscale` on the Terraform Registry.

References:

- HashiCorp provider publishing docs: https://developer.hashicorp.com/terraform/registry/providers/publishing
- HashiCorp provider documentation docs: https://developer.hashicorp.com/terraform/registry/providers/docs
- DNScale registry page: https://registry.terraform.io/providers/dnscaleou/dnscale

## Current Release Setup

- GitHub repository: `dnscaleou/terraform-provider-dnscale`
- Registry provider source: `dnscaleou/dnscale`
- Provider server address: `registry.terraform.io/dnscaleou/dnscale`
- Release automation: `.github/workflows/release.yml`
- Build/sign/upload config: `.goreleaser.yml`
- Registry manifest: `terraform-registry-manifest.json`
- Registry protocol: `6.0` because this provider uses Terraform Plugin Framework protocol 6.

The release workflow runs on tags matching `v*`. GoReleaser builds zip archives, includes the registry manifest, writes SHA256 checksums, signs the checksum file with GPG, and creates the GitHub release assets that the Terraform Registry ingests.

## One-Time Registry Requirements

These should already be configured for the existing `dnscaleou/dnscale` provider, but verify them before the next release if publishing fails:

- The GitHub repo must be public and named `terraform-provider-dnscale`.
- The Terraform Registry must have the signing public key registered under the `dnscaleou` namespace.
- GitHub repository secrets must exist:
  - `GPG_PRIVATE_KEY`: ASCII-armored private key exported with `gpg --armor --export-secret-keys <key-id-or-email>`.
  - `PASSPHRASE`: passphrase for the private key.
- The Registry webhook should be present on the GitHub repository. If releases stop appearing, use the provider settings page in the Registry to resync the webhook.

HashiCorp requires valid SemVer tags prefixed with `v`, for example `v1.0.1`. Do not replace, retag, or mutate an already published version; publish a new version instead.

## Release Checklist

1. Start from a clean provider worktree, except for any unrelated user-local files that should not be released.

   ```bash
   git status --short
   ```

2. Run formatting and tests.

   ```bash
   gofmt -w .
   GOCACHE=$(pwd)/.gocache go test ./...
   rm -rf .gocache
   ```

   Acceptance tests require a live API key and will skip without `DNSCALE_API_KEY` where implemented:

   ```bash
   TF_ACC=1 DNSCALE_API_KEY=... go test ./... -run TestAcc -count=1
   ```

3. Commit the release changes.

   ```bash
   git add .
   git commit -m "chore: prepare terraform provider release"
   ```

4. Choose the next SemVer tag. For bug fixes, use the next patch version.

   ```bash
   git tag -a v1.0.1 -m "v1.0.1"
   ```

5. Push the commit and tag.

   ```bash
   git push origin main
   git push origin v1.0.1
   ```

6. Watch the GitHub Actions release workflow. It should create a GitHub release with:

   - `terraform-provider-dnscale_<version>_<os>_<arch>.zip`
   - `terraform-provider-dnscale_<version>_manifest.json`
   - `terraform-provider-dnscale_<version>_SHA256SUMS`
   - `terraform-provider-dnscale_<version>_SHA256SUMS.sig`

7. Verify the Terraform Registry ingested the new version:

   ```bash
   terraform init -upgrade
   terraform providers
   ```

   Also check the public page: https://registry.terraform.io/providers/dnscaleou/dnscale

## Local Dry Run

Use this before pushing a tag if release config changes:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

The snapshot command does not publish to GitHub or the Registry, but it validates that GoReleaser can build the expected archives locally.
