<?php
// Owned differential fixture: the PHP 8.4-only array_all subset accepts one
// array and a literal name of a declared global callback with exactly two
// required, untyped parameters and an explicit bool return type. It is not a
// general PHP callable implementation.

function mutate_nested_all($value, $key): bool {
    $value['changed'] = 'inside-callback';
    return is_array($value) && $key === 'nested';
}

function token_is_ok($value, $key): bool {
    echo 'token:' . $key . ':' . strtok($value, ',') . "\n";
    return strtok(',') === 'ok';
}

// Empty input invokes no callback and is true. The ordinary all-match and
// false cases expose normalized integer keys, string keys, and stopping at
// the first false callback result.
$empty = ArRaY_AlL([], 'visit_all');
echo 'empty:' . ($empty === true ? 'true' : 'wrong') . "\n";

$all = array_all(["2" => 'ok', 'name' => 'ok'], 'ViSiT_AlL');
echo 'all:' . ($all === true ? 'true' : 'wrong') . "\n";

$first_false = ARRAY_ALL(['bad', 'late'], 'visit_all');
echo 'first-false:' . ($first_false === false ? 'false' : 'wrong') . "\n";

$middle_false = array_all([10 => 'ok', 'name' => 'bad', 'late' => 'dead'], 'visit_all');
echo 'middle-false:' . ($middle_false === false ? 'false' : 'wrong') . "\n";

// Parameters are copied at the callback boundary. The nested caller value
// remains untouched even though the callback writes its local array copy.
$nested = ['nested' => ['original' => 'kept']];
$nested_all = array_all($nested, 'MuTaTe_NeStEd_AlL');
echo 'nested:' . ($nested_all ? 'true' : 'wrong') . ':'
    . (isset($nested['nested']['changed']) ? 'changed' : 'clean') . ':'
    . $nested['nested']['original'] . "\n";

// Stateful callbacks retain the one request execution. The second element
// fails, so the third one is never visited.
$tokens = array_all(['first,ok', 'second,stop', 'third,ok'], 'ToKeN_Is_Ok');
echo 'tokens:' . ($tokens === false ? 'false' : 'wrong') . "\n";

// Both PHP 8.4 predicates compose as ordinary bool expressions.
$composition = array_all(['ok', 'ok'], 'visit_all')
    && array_any(['bad', 'ok'], 'ViSiT_AlL');
echo 'composition:' . ($composition ? 'yes' : 'wrong') . "\n";

// Forward resolution must work even when the named declaration follows calls.
function ViSiT_AlL($value, $key): BoOl {
    echo 'visit:' . gettype($key) . ':' . $key . ':' . $value . "\n";
    return $value === 'ok';
}
