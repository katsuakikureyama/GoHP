<?php
// Owned differential fixture: bounded isset() accepts defined bare variables
// and one-level indexed variables. Multiple operands are evaluated left to
// right and short-circuit at the first null/missing operand.
function bool_word($value) {
    if ($value) {
        return 'T';
    }
    return 'F';
}

function return_isset($value) {
    return IsSeT($value);
}

function multi_isset($first, $second, $third) {
    return isset($first, $second, $third);
}

function safe_probe($key, $items) {
    return isset($key, $items[$key]);
}

$null_value = null;
$false_value = false;
$zero_value = 0;
$empty_value = '';
$array_value = [];
$callable_value = fn($value) => $value;
$items = ['present' => false, 'null' => null, 'zero' => 0, 'empty' => ''];
$present_key = 'present';
$missing_key = 'missing';
$poison = 'not-an-array';

$logical = isset($false_value, $zero_value) && isset($empty_value);
$ternary = isset($null_value, $false_value) ? 'wrong' : 'safe';
$returned = return_isset($false_value);
$short_circuit = safe_probe(null, $poison);

echo 'bare=' . bool_word(isset($null_value)) . ':' . bool_word(isset($false_value)) . ':' . bool_word(isset($zero_value)) . ':' . bool_word(isset($empty_value)) . ':' . bool_word(isset($array_value)) . ':' . bool_word(isset($callable_value)) . "\n";
echo 'multi=' . bool_word(multi_isset($false_value, $zero_value, $empty_value)) . ':' . bool_word(isset($false_value, $null_value, $empty_value)) . ':' . bool_word(isset($false_value, $items[$present_key], $empty_value)) . "\n";
echo 'index=' . bool_word(isset($items[$present_key])) . ':' . bool_word(isset($items['null'])) . ':' . bool_word(isset($items[$missing_key])) . ':' . bool_word(isset($items['zero'])) . ':' . bool_word(isset($items['empty'])) . "\n";
echo 'contexts=' . bool_word($logical) . ':' . $ternary . ':' . bool_word($returned) . ':' . bool_word($short_circuit) . "\n";
