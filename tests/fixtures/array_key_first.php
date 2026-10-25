<?php
// Owned differential fixture: array_key_first() observes the first insertion-
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

function forwarded_first_key(array $items, callable $callback) {
    return $callback($items);
}

function mapped_first_key(array $items) {
    return array_key_first($items);
}

// The empty guard is intentional: the element access must only occur after a
// non-empty array has supplied a real first key, including a null or false
// first value.
function first_or_null(array $items) {
    if (empty($items)) {
        return null;
    }
    return $items[array_key_first($items)];
}

$empty = [];
$zero = [0 => 'zero'];
$blank = ['' => 'blank'];
$sparse = [-3 => 'negative', 8 => 'eight', 'tail' => 'tail'];
$non_sorted = [9 => 'nine', -2 => 'minus', 4 => 'four'];
$negative_null = [-5 => null];
$canonical_numeric_null = ['2' => null];
$leading_numeric_null = ['02' => null];

// Canonical numeric-string keys replace their original entry without moving
// it; the already-later key must remain last.
$normalized = [1 => 'first', 'tail' => 'tail'];
$normalized['1'] = 'replaced';

// Updating the first entry and appending after a later numeric key preserves
// both the original first key and the final key. Appending to an empty array
// establishes integer key zero as its first key.
$updated = ['first' => 'before', 4 => 'four'];
$updated['first'] = 'after';
$updated[] = 'appended';
$from_empty = [];
$from_empty[] = 'only';

$strict_zero = array_key_first($zero) === 0;
$strict_blank = array_key_first($blank) === '';
$strict_empty = array_key_first($empty) === null;
$strict_sparse = array_key_first($sparse) === -3;
$marker = array_key_first(marked_array($updated));
$forwarded = forwarded_first_key($sparse, fn($items) => array_key_first($items));
$mapped = array_map('mapped_first_key', [[10 => 'ten'], ['name' => false], []]);

$null_first = ['first' => null, 'tail' => true];
$false_first = ['first' => false, 'tail' => true];
$nested_first = ['first' => ['inside' => 'value'], 'tail' => true];
$empty_value = first_or_null($empty);
$null_value = first_or_null($null_first);
$false_value = first_or_null($false_first);
$nested_value = first_or_null($nested_first);

echo 'basic=' . key_word(ArRaY_KeY_FiRsT($zero)) . ':' . key_word(array_key_first($blank)) . ':' . key_word(array_key_first($sparse)) . ':' . key_word(array_key_first($non_sorted)) . ':' . key_word(array_key_first($empty)) . "\n";
echo 'key-types=' . key_word(array_key_first($negative_null)) . ':' . key_word(array_key_first($canonical_numeric_null)) . ':' . key_word(array_key_first($leading_numeric_null)) . "\n";
echo 'strict=' . bool_word($strict_zero) . ':' . bool_word($strict_blank) . ':' . bool_word($strict_empty) . ':' . bool_word($strict_sparse) . "\n";
echo 'normalized=' . key_word(array_key_first($normalized)) . ':' . key_word(array_key_last($normalized)) . ':' . $normalized[1] . ':' . $normalized['tail'] . "\n";
echo 'updated=' . key_word(array_key_first($updated)) . ':' . key_word(array_key_last($updated)) . ':' . key_word(array_key_first($from_empty)) . ':' . count($updated) . "\n";
echo 'marker=' . key_word($marker) . "\n";
echo 'callback=' . key_word($forwarded) . ':' . key_word($mapped[0]) . ':' . key_word($mapped[1]) . ':' . key_word($mapped[2]) . "\n";
echo 'values=' . bool_word(is_null($empty_value)) . ':' . bool_word(is_null($null_value)) . ':' . bool_word($false_value === false) . ':' . $nested_value['inside'] . ':' . key_word(array_key_first($null_first)) . ':' . key_word(array_key_first($false_first)) . ':' . key_word(array_key_first($nested_first)) . "\n";
