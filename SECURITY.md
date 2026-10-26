# Security (proposal)

This compiler and generated runtime are experimental and are not a security
sandbox. PHP input, upstream build scripts, Composer hooks and generated binaries
are untrusted. Never execute customer input with production secrets or network/DB
access. Container tests use only owned fixtures, no network, dropped capabilities,
read-only root filesystem and temporary writable outputs. Docker alone is not a
sufficient multi-tenant hostile-build boundary; use stronger isolated workers,
resource quotas and reviewed artifact promotion before a service deployment.

## PHP oracle maintenance

The active native compatibility oracles are PHP 8.3.36 or newer in the 8.3
series and PHP 8.4.27 or newer in the 8.4 series. Each target has its own resolved
immutable image digest and run evidence; do not substitute a floating tag or
transfer evidence between targets. The [active version policy](docs/compatibility.md#active-php-version-policy)
and [CI matrix](.github/workflows/ci.yml) define the current admission rules.
PHP release notices and the [supported-versions policy](https://www.php.net/supported-versions.php)
must be reviewed when updating the approved oracle image or when deciding to
add another series.

This version floor does **not** make PHP input, the compiler, generated code,
container images or dependencies secure. It does not replace CVE triage,
digest review, upstream support/EOL review, isolation, or normal release
approval. Image-digest updates and EOL review are tracked maintenance work;
this repository has no automatic image-update or automatic support-promotion
mechanism.

No eval, shell execution, include/autoload or Composer hook support exists. There
are no production migrations, payments or deployment commands. Report security
issues privately to the repository owner through GitHub's available private
security reporting channel; if unavailable, request a private channel without
publishing exploit details. Licensing and public release remain owner decisions.
