<?php
// Owned differential fixture: array_key_last() observes the final insertion-
// order key without changing the array or inspecting its value.
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
    if (is_null($key)) {
        return 'N';
    }
    return 'X';
}

function marked_array(array $items): array {
    echo '<marked>';
    return $items;
}

function forwarded_last_key(array $items, callable $callback) {
    return $callback($items);
}

function mapped_last_key(array $items) {
    return array_key_last($items);
}

function last_or_null(array $items) {
    if (empty($items)) {
        return null;
    }
    return $items[array_key_last($items)];
}

$empty = [];
$zero = [0 => 'zero'];
$blank = ['' => 'blank'];
$sparse = [-3 => 'negative', 8 => 'eight', 'tail' => 'tail'];
$non_sorted = [9 => 'nine', -2 => 'minus', 4 => 'four'];
$negative_null = [-5 => null];
$canonical_numeric_null = ['2' => null];
$leading_numeric_null = ['02' => null];

// Canonical numeric-string keys replace their existing entry; replacement
// must not move that key after the already-later string key.
$normalized = [1 => 'first', 'tail' => 'tail'];
$normalized['1'] = 'replaced';

// Updating an earlier entry keeps its position; append selects the next
// numeric key and becomes the final insertion-order entry.
$updated = ['first' => 'before', 4 => 'four'];
$updated['first'] = 'after';
$keys_before = array_keys($updated);
$updated[] = 'appended';
$last_updated = array_key_last($updated);
$keys_after = array_keys($updated);

$strict_zero = array_key_last($zero) === 0;
$strict_blank = array_key_last($blank) === '';
$strict_tail = array_key_last($sparse) === 'tail';
$strict_unsorted = array_key_last($non_sorted) === 4;
$strict_empty = array_key_last($empty) === null;
$marker = array_key_last(marked_array($updated));
$forwarded = forwarded_last_key($sparse, fn($items) => array_key_last($items));
$mapped = array_map('mapped_last_key', [[10 => 'ten'], ['name' => false], []]);

$null_tail = ['first' => true, 'tail' => null];
$false_tail = ['first' => true, 'tail' => false];
$nested_tail = ['first' => true, 'tail' => ['inside' => 'value']];
$empty_value = last_or_null($empty);
$null_value = last_or_null($null_tail);
$false_value = last_or_null($false_tail);
$nested_value = last_or_null($nested_tail);

echo 'basic=' . key_word(ArRaY_KeY_LaSt($zero)) . ':' . key_word(array_key_last($blank)) . ':' . key_word(array_key_last($sparse)) . ':' . key_word(array_key_last($non_sorted)) . ':' . key_word(array_key_last($empty)) . "\n";
echo 'key-types=' . key_word(array_key_last($negative_null)) . ':' . key_word(array_key_last($canonical_numeric_null)) . ':' . key_word(array_key_last($leading_numeric_null)) . "\n";
echo 'strict=' . bool_word($strict_zero) . ':' . bool_word($strict_blank) . ':' . bool_word($strict_tail) . ':' . bool_word($strict_unsorted) . ':' . bool_word($strict_empty) . "\n";
echo 'normalized=' . key_word(array_key_last($normalized)) . ':' . $normalized[1] . ':' . $normalized['tail'] . "\n";
echo 'updated=' . key_word($last_updated) . ':' . count($updated) . ':' . $keys_before[0] . ':' . $keys_before[1] . ':' . $keys_after[0] . ':' . $keys_after[1] . ':' . $keys_after[2] . "\n";
echo 'marker=' . key_word($marker) . "\n";
echo 'callback=' . key_word($forwarded) . ':' . key_word($mapped[0]) . ':' . key_word($mapped[1]) . ':' . key_word($mapped[2]) . "\n";
echo 'values=' . bool_word(is_null($empty_value)) . ':' . bool_word(is_null($null_value)) . ':' . bool_word($false_value === false) . ':' . $nested_value['inside'] . ':' . key_word(array_key_last($null_tail)) . ':' . key_word(array_key_last($false_tail)) . ':' . key_word(array_key_last($nested_tail)) . "\n";
