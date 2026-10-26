<?php
// Owned differential fixture: array_slice() selects an insertion-ordered
// range. Integer keys are reindexed by default; string keys stay attached,
// while preserve_keys retains every selected key.
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

function marker_array($items) {
    echo '<array>';
    return $items;
}

function marker_offset($value) {
    echo '<offset>';
    return $value;
}

function marker_length($value) {
    echo '<length>';
    return $value;
}

function marker_flag($value) {
    echo '<flag>';
    return $value;
}

function dynamic_slice($items, $offset, $length, $preserve): array {
    return array_slice($items, $offset, $length, $preserve);
}

// The canonical numeric string '2' is an integer key, whereas '02' remains
// a string key. The maximum key is used only by slices that either reindex it
// or exclude it before appending.
$mixed = [
    -5 => 'minus-five',
    'name' => 'named',
    8 => 'eight',
    '2' => 'canonical-two',
    '02' => 'leading-two',
    9223372036854775807 => 'maximum',
    'nested' => ['inside' => 'source'],
];

// Two, three, and four argument forms cover omitted, explicit, and preserved
// keys. The explicit null length means "through the end".
$two = ArRaY_SlIcE($mixed, 2);
$three = array_slice($mixed, -5, 3);
$four = array_slice($mixed, -6, 4, true);
$null_length = array_slice($mixed, 2, null);

// Positive and negative out-of-range offsets clamp to the end and start.
// Positive, zero, negative, null, and omitted lengths all take distinct paths.
$after_end = array_slice($mixed, 99);
$before_start = array_slice($mixed, -99);
$zero_length = array_slice($mixed, 1, 0);
$negative_length = array_slice($mixed, 1, -2);

// The signed bounds must be handled without negating the minimum integer.
$minimum = -9223372036854775807 + -1;
$maximum = 9223372036854775807;
$minimum_offset = array_slice($mixed, $minimum, 2);
$maximum_offset = array_slice($mixed, $maximum, 2);
$minimum_length = array_slice($mixed, 0, $minimum);
$maximum_length = array_slice($mixed, 1, $maximum);

// Results are independent arrays, including nested values. Extract-and-rebind
// keeps this fixture within the supported assignment syntax.
$nested_source = ['before' => 'B', 'nested' => ['inside' => 'source'], 5 => 'five'];
$nested_result = array_slice($nested_source, 1, 1, true);
$nested_copy = $nested_result['nested'];
$nested_copy['inside'] = 'changed';
$nested_result['nested'] = $nested_copy;

// Preserved negative integer keys append after the selected largest key in
// PHP 8.3. A maximum key is deliberately excluded before the preserved result
// appends, so it cannot poison the result's next-key metadata.
$negative_source = [-5 => 'minus-five', -10 => 'minus-ten', 'name' => 'named'];
$negative_preserved = array_slice($negative_source, 0, null, true);
$negative_preserved[] = 'tail';
$maximum_source = [9223372036854775807 => 'maximum', 'name' => 'named'];
$maximum_reindexed = array_slice($maximum_source, 0);
$maximum_reindexed[] = 'tail';
$maximum_excluded = array_slice($maximum_source, 1, null, true);
$maximum_excluded[] = 'tail';

// Each argument expression is observed exactly once in positional order.
$marked = array_slice(
    marker_array($mixed),
    marker_offset(1),
    marker_length(3),
    marker_flag(true)
);

// Untyped parameters exercise all four runtime argument positions.
$dynamic = dynamic_slice($mixed, -4, null, false);
$empty = array_slice([], $minimum, null, true);
$empty[] = 'first';

$ignored = dump_entries('input', $mixed);
$ignored = dump_entries('empty-append', $empty);
$ignored = dump_entries('two', $two);
$ignored = dump_entries('three', $three);
$ignored = dump_entries('four', $four);
$ignored = dump_entries('null-length', $null_length);
echo 'bounds=' . count($after_end) . ':' . count($before_start) . ':' . count($zero_length) . ':' . count($negative_length) . "\n";
echo 'extremes=' . count($minimum_offset) . ':' . count($maximum_offset) . ':' . count($minimum_length) . ':' . count($maximum_length) . "\n";
$ignored = dump_entries('before-start', $before_start);
$ignored = dump_entries('negative-length', $negative_length);
$ignored = dump_entries('minimum-offset', $minimum_offset);
$ignored = dump_entries('maximum-length', $maximum_length);
echo 'nested=' . $nested_source['nested']['inside'] . ':' . $nested_result['nested']['inside'] . ':' . $nested_source[5] . "\n";
$ignored = dump_entries('negative-preserved', $negative_preserved);
$ignored = dump_entries('maximum-reindexed', $maximum_reindexed);
$ignored = dump_entries('maximum-excluded', $maximum_excluded);
$ignored = dump_entries('marked', $marked);
$ignored = dump_entries('dynamic', $dynamic);
