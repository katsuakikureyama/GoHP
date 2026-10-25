<?php
// Owned differential fixture: array_reverse() returns a distinct array in
// reverse insertion order. Its optional flag decides whether integer keys are
// retained; string keys always remain attached to their values.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function key_word($key) {
    if (is_int($key)) {
        return 'I:' . $key;
    }
    if (is_string($key)) {
        return 'S:' . $key;
    }
    return 'X';
}

function dump_entries($label, array $items) {
    echo $label . '=';
    foreach ($items as $key => $value) {
        echo key_word($key) . ':';
        if (is_array($value)) {
            echo 'A' . count($value);
        } else {
            echo $value;
        }
        echo ',';
    }
    echo "\n";
}

function marker_array(array $items): array {
    echo '<array>';
    return $items;
}

function marker_flag($value): bool {
    echo '<flag>';
    return $value;
}

function dynamic_reverse(array $items, $preserve): array {
    return array_reverse($items, $preserve);
}

function forward_reverse(array $items, $preserve, callable $callback): array {
    return $callback($items, $preserve);
}

function reverse_default(array $items): array {
    return array_reverse($items);
}

$empty = [];
// The canonical numeric string '2' is an integer key, whereas '02' remains
// a string key. The ordering deliberately interleaves both key categories.
$mixed = [-5 => 'minus', 'name' => 'named', 8 => 'eight', '2' => 'two', '02' => 'leading'];
$default = ArRaY_ReVeRsE($mixed);
$false = array_reverse($mixed, false);
$true = array_reverse($mixed, true);

// Reversing a result must not mutate the input, and modifying a nested result
// must not change the original nested array.
$nested_source = ['head' => ['inside' => 'before'], -2 => 'negative', 4 => 'four'];
$nested_result = array_reverse($nested_source, true);
$nested_copy = $nested_result['head'];
$nested_copy['inside'] = 'changed';
$nested_result['head'] = $nested_copy;

// A preserved sequence of negative integer keys receives max-key-plus-one on
// append in PHP 8.3. The unpreserved sequence is reindexed before it appends.
$negative = [-5 => 'minus-five', -10 => 'minus-ten'];
$negative_true = array_reverse($negative, true);
$negative_true[] = 'preserved-tail';
$negative_false = array_reverse($negative, false);
$negative_false[] = 'reindexed-tail';

// Retaining PHP_INT_MAX would make a following append invalid. The false
// branch instead renumbers that integer key, so it can append safely.
$maximum = [9223372036854775807 => 'maximum', 'name' => 'named'];
$maximum_true = array_reverse($maximum, true);
$maximum_false = array_reverse($maximum, false);
$maximum_false[] = 'tail';

// Both arguments are expressions with observable effects and must be
// evaluated exactly once, left to right.
$marked = array_reverse(marker_array($mixed), marker_flag(true));

// Array and untyped flag parameters exercise runtime values through a function and a
// capture-free callback; array_map composes the default-flag path as well.
$dynamic_true = dynamic_reverse($mixed, true);
$callback_false = forward_reverse($mixed, false, fn($items, $preserve) => array_reverse($items, $preserve));
$mapped = array_map('reverse_default', [[3 => 'three', 'key' => 'value'], [-1 => 'minus', 5 => 'five']]);

echo 'empty=' . count(array_reverse($empty)) . ':' . count(array_reverse($empty, true)) . "\n";
$ignored = dump_entries('default', $default);
$ignored = dump_entries('false', $false);
$ignored = dump_entries('true', $true);
$ignored = dump_entries('input', $mixed);
echo 'nested=' . $nested_source['head']['inside'] . ':' . $nested_result['head']['inside'] . ':' . $nested_source[-2] . ':' . $nested_result[-2] . "\n";
$ignored = dump_entries('negative-true', $negative_true);
$ignored = dump_entries('negative-false', $negative_false);
$ignored = dump_entries('maximum-true', $maximum_true);
$ignored = dump_entries('maximum-false', $maximum_false);
$ignored = dump_entries('marked', $marked);
$ignored = dump_entries('dynamic-true', $dynamic_true);
$ignored = dump_entries('callback-false', $callback_false);
$ignored = dump_entries('mapped-first', $mapped[0]);
$ignored = dump_entries('mapped-second', $mapped[1]);
