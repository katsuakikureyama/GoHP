# Last insertion-order array key

The matching first-key operation has its own
[`array_key_first` contract](array-key-first-boundary.md).

`array_key_last(array)` accepts exactly one actual array, with case-insensitive
direct call spelling. It returns the last outer entry's stored int64 or string
key, or canonical null for an empty array. Zero and an empty-string key are not
null. This follows [PHP's array_key_last contract](https://www.php.net/manual/en/function.array-key-last.php)
within the existing GoHP array domain.

The operation does not sort or renumber keys. Replacing an existing key preserves
its insertion position; inserting a new key or appending an element makes that
new entry last. Existing numeric-string normalization and PHP 8.3 negative-key
append rules apply. Values, including null, nested arrays and generated objects,
are not inspected or converted. Input and next-index metadata are not modified.
The argument is evaluated once.

Known nonarray arguments, wrong arity and reserved global redeclaration fail
before generation. Dynamic nonarrays fail at `rt.ArrayKeyLast` without coercion.
The semantic model has no int|string|null union, so the result retains the
existing unknown kind and downstream runtime guards. An empty result cannot be
used as an array index: guard with bounded `empty($array)` or `is_null($key)`
before reading. This does not add narrowing or weak/null key coercion.

Native evidence compares owned fixture stdout and normal exits on PHP 8.3
>=8.3.36: typed keys, empty/zero distinction, order, replacement/append, guarded
last-value reads, single evaluation and callbacks. Runtime tests include integer
extremes, binary string keys and metadata preservation. General PHP error/exception
equivalence, named arguments and builtin callable lookup are excluded. Internal
array-pointer APIs are not implemented; no cursor interoperability claim is made.

The catalog remains partial. Gossip's original-source pins/results are unchanged;
adoption requires separate evidence. PHP 8.4+ remains a separate target decision.
`gohp.cli.generate` stays WebMCP-excluded local tooling, with no new server
operation, route, authentication, dependency, release or deployment.
