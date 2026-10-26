<?php
// Owned differential fixture: the PHP 8.4-only array_any subset accepts one
// array and a literal name of a declared global callback with exactly two
// required, untyped parameters and an explicit bool return type. It is not a
// general PHP callable implementation.

function mutate_nested($value, $key): bool {
    $value['changed'] = 'inside-callback';
    return is_array($value) && $key === 'nested';
}

function token_has_hit($value, $key): bool {
    echo 'token:' . $key . ':' . strtok($value, ',') . "\n";
    return strtok(',') === 'hit';
}

// Empty input invokes no callback. The ordinary no-match and hit cases expose
// order, normalized integer keys, string keys, and the first-match stop.
$empty = ArRaY_AnY([], 'visit_match');
echo 'empty:' . ($empty === false ? 'false' : 'wrong') . "\n";

$none = array_any(["2" => 'miss', 'name' => 'also-miss'], 'ViSiT_MaTcH');
echo 'none:' . ($none === false ? 'false' : 'wrong') . "\n";

$first = ARRAY_ANY(['hit', 'late'], 'visit_match');
echo 'first:' . ($first ? 'true' : 'wrong') . "\n";

$middle = array_any([10 => 'miss', 'name' => 'hit', 'late' => 'dead'], 'visit_match');
echo 'middle:' . ($middle === true ? 'true' : 'wrong') . "\n";

// Parameters are copied at the callback boundary. The nested caller value
// remains untouched even though the callback writes its local array copy.
$nested = ['nested' => ['original' => 'kept']];
$nested_any = array_any($nested, 'MuTaTe_NeStEd');
echo 'nested:' . ($nested_any ? 'true' : 'wrong') . ':'
    . (isset($nested['nested']['changed']) ? 'changed' : 'clean') . ':'
    . $nested['nested']['original'] . "\n";

// Stateful callbacks retain the one request execution. The second element
// matches and therefore the third is never visited.
$tokens = array_any(['first,miss', 'second,hit', 'third,hit'], 'ToKeN_HaS_HiT');
echo 'tokens:' . ($tokens === true ? 'true' : 'wrong') . "\n";

// The boolean result is available in ordinary supported expressions.
echo 'expression:' . ($first && !$none ? 'yes' : 'wrong') . "\n";

// Forward resolution must work even when the named declaration follows calls.
function ViSiT_MaTcH($value, $key): BoOl {
    echo 'visit:' . gettype($key) . ':' . $key . ':' . $value . "\n";
    return $value === 'hit';
}
