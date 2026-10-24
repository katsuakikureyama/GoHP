# Security (proposal)

This compiler and generated runtime are experimental and are not a security
sandbox. PHP input, upstream build scripts, Composer hooks and generated binaries
are untrusted. Never execute customer input with production secrets or network/DB
access. Container tests use only owned fixtures, no network, dropped capabilities,
read-only root filesystem and temporary writable outputs. Docker alone is not a
sufficient multi-tenant hostile-build boundary; use stronger isolated workers,
resource quotas and reviewed artifact promotion before a service deployment.

No eval, shell execution, include/autoload or Composer hook support exists. There
are no production migrations, payments or deployment commands. Report security
issues privately to the repository owner through GitHub's available private
security reporting channel; if unavailable, request a private channel without
publishing exploit details. Licensing and public release remain owner decisions.
