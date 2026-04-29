# How to Prepare a Branch for a New Dex Version

Whenever a new Dex version needs to be shipped with GitOps, we need to prepare the corresponding version branch and release tag in the midstream Dex repository.

These branches and tags are referenced by our downstream build system in the release repository:

```text
https://github.com/rh-gitops-midstream/release
```

The midstream Dex repository is used to track the upstream Dex version we ship and to provide a place where we can apply downstream fixes, such as CVE fixes, when required.

## 1. Clone the midstream Dex repository

```bash
git clone https://github.com/rh-gitops-midstream/dex.git
cd dex
```

## 2. Add the upstream Dex remote

```bash
git remote add upstream https://github.com/dexidp/dex.git
```

If the `upstream` remote already exists, verify that it points to the correct upstream repository:

```bash
git remote -v
```

## 3. Create a version branch from the upstream tag

We create a branch from the upstream Dex tag that we plan to ship.

This branch is required because we may need to apply fixes in the future for this specific Dex version. For example, upstream may fix a CVE only in a later release and may not accept patches for an older version that we still support downstream.

Fetch the upstream tags:

```bash
git fetch upstream --tags
```

Create a new branch from the upstream tag:

```bash
git checkout v2.43.1
git switch -c v2.43.1
```

Push the version branch to the midstream repository:

```bash
git push origin refs/heads/v2.43.1
```

Alternatively, the branch can be pushed more explicitly as:

```bash
git push origin v2.43.1:v2.43.1
```

## 4. Create a release tag from the version branch

Our downstream build system references tags, not just branches. After creating the version branch, create a release tag from that branch.

The tag should use the following format:

```text
<upstream-dex-version>-<release-number>
```

For example:

```bash
git tag v2.43.1-1
git push origin v2.43.1-1
```

The `-1` suffix identifies the first downstream release tag for that Dex version.

## 5. Incrementing release tags

If we need to cut another tag for the same Dex version, increment the suffix.

For example:

```text
v2.43.1-1
v2.43.1-2
v2.43.1-3
```

This is typically done when we need to rebuild the same Dex version with additional downstream changes, such as a dependency update or CVE fix.

If the Dex Y-stream or Z-stream version changes, create a new version branch and a new release tag by following the same process again.

For example, moving from `v2.43.1` to `v2.43.2` should use a new branch and tag:

```text
branch: v2.43.2
tag:    v2.43.2-1
```

# How to Address a CVE in Dex

To address a CVE in Dex, the fix must be brought into the midstream Dex repository. The approach depends on where the fix is available and whether we can safely move to a newer Dex version.

## Preferred approach: move to a fixed Dex version

If the CVE is fixed in a newer Dex release and that version is acceptable for the target GitOps release, prefer updating to the fixed Dex version.

In that case:

1. Create a new branch from the fixed upstream Dex tag.
2. Create a new downstream release tag with the `-1` suffix.
3. Update the downstream release repository to reference the new Dex tag.

## Keep Dex aligned with upstream Argo CD when possible

We try to keep the Dex Y-stream aligned with the Dex version used by upstream Argo CD.

If a CVE is fixed only in a later Dex Y-stream, avoid bumping the Dex Y-stream unless it is required. A Y-stream bump may introduce behavioral changes or compatibility concerns.

Z-stream bumps are generally safer and can usually be taken with lower risk.

## Alternative approach: patch the supported midstream version

If moving to a newer Dex version is not possible, update the affected dependency in the supported midstream version branch.

This should be done by updating the dependency in `go.mod` if the CVE can be resolved through a dependency bump.

After applying the fix:

```bash
go mod tidy
```

Then run the required build and test validation before cutting a new downstream release tag.

For example, if the current supported branch is `v2.43.1` and the latest downstream tag is `v2.43.1-1`, the CVE fix should be tagged as:

```text
v2.43.1-2
```

## Recommended decision flow

```text
Dex CVE found
        |
        v
Is the CVE fixed in a newer Dex Z-stream that is safe to consume?
        |
        +-- Yes --> Create new branch from upstream tag and cut a new -1 release tag
        |
        +-- No
             |
             v
Is the CVE fixed only in a newer Dex Y-stream?
        |
        +-- Yes --> Check compatibility with upstream Argo CD before bumping
        |
        +-- No / Not suitable
             |
             v
Backport the fix or update the affected dependency in the supported midstream branch
        |
        v
Cut a new downstream release tag by incrementing the suffix
```

## Summary

For every new Dex version shipped with GitOps, create a version branch from the upstream Dex tag and then create a downstream release tag using the `<version>-<number>` format.

For CVE fixes, prefer consuming a fixed Dex Z-stream when possible. Avoid Dex Y-stream bumps unless required, because we try to stay aligned with the Dex version used by upstream Argo CD. If a version bump is not suitable, backport the fix or update the affected dependency in the supported midstream branch and cut a new incremented release tag.
