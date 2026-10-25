# Security (proposal)

This compiler and generated runtime are experimental and are not a security
sandbox. PHP input, upstream build scripts, Composer hooks and generated binaries
are untrusted. Never execute customer input with production secrets or network/DB
access. Container tests use only owned fixtures, no network, dropped capabilities,
read-only root filesystem and temporary writable outputs. Docker alone is not a
sufficient multi-tenant hostile-build boundary; use stronger isolated workers,
resource quotas and reviewed artifact promotion before a service deployment.

## PHP oracle maintenance

The active native compatibility oracle is PHP 8.3.36 or newer in the 8.3
series. Its resolved immutable image digest is tracked by the test policy and
run evidence; do not substitute a floating tag without review. PHP's
[8.3.36 release](https://www.php.net/releases/8_3_36.php) is a security release,
and the [supported-versions policy](https://www.php.net/supported-versions.php)
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
