<?php
// Owned differential fixture: array_merge() accepts zero or more arrays,
// reindexes integer keys, and keeps string keys in their original insertion
// position when a later argument replaces their value.
function key_word($key) {
    if (is_int($key)) {
        return 'I:' . $key;
    }
    if (is_string($key)) {
        return 'S:' . $key;
    }
    return 'X';
}

function value_word($value) {
    if (is_array($value)) {
        return 'A' . count($value);
    }
    if (is_null($value)) {
        return 'N';
    }
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return $value;
}

function dump_entries($label, array $items) {
    echo $label . '=';
    foreach ($items as $key => $value) {
        echo key_word($key) . ':' . value_word($value) . ',';
    }
    echo "\n";
}

function marker_first(array $items): array {
    echo '<first>';
    return $items;
}

function marker_second(array $items): array {
    echo '<second>';
    return $items;
}

function marker_third(array $items): array {
    echo '<third>';
    return $items;
}

function dynamic_merge($first, $second, $third): array {
    return array_merge($first, $second, $third);
}

function forward_merge(array $first, array $second, callable $callback): array {
    return $callback($first, $second);
}

function merge_callback(array $items): array {
    return array_merge(['prefix' => 'P'], $items, [9 => 'nine']);
}

// Zero and one argument calls both return a fresh array. Editing the nested
// result after the one-argument call cannot alter the source.
$empty = [];
$zero = ArRaY_MeRgE();
$one_source = ['nested' => ['inside' => 'source'], -2 => 'negative', 'name' => 'one'];
$one = array_merge($one_source);
$one_nested = $one['nested'];
$one_nested['inside'] = 'changed';
$one['nested'] = $one_nested;

// Numeric keys, including the canonical numeric string '2', are renumbered.
// The leading-zero key '02' remains a distinct string key. The maximum key is
// intentionally safe here because array_merge renumbers it before append.
$first = [
    'first' => 'F',
    'duplicate' => 'old',
    'after' => 'A',
    -5 => 'minus-five',
    8 => 'eight',
    '2' => 'canonical-two',
    '02' => 'leading-two',
    9223372036854775807 => 'maximum',
];
$second = [
    -5 => 'duplicate-integer',
    'duplicate' => null,
    'second' => false,
    'nested' => ['from' => 'second', 'discarded' => 'old'],
];
$third = [
    'duplicate' => 'final',
    -10 => 'minus-ten',
    'nested' => ['from' => 'third'],
    'third' => 'T',
];
$two = array_merge($first, $second);
$merged = array_merge($first, $second, $third);
$merged[] = 'tail';

// The nested value is replaced as a whole, never recursively merged. Rebind a
// changed copy to prove both the result and the third input have independent
// outer/nested arrays.
$merged_nested = $merged['nested'];
$merged_nested['from'] = 'changed';
$merged['nested'] = $merged_nested;

// Each positional expression is observed once in left-to-right order.
$marked = array_merge(
    marker_first(['left' => 'L', 7 => 'seven']),
    marker_second(['middle' => 'M', -1 => 'minus']),
    marker_third(['right' => 'R'])
);

// All three untyped arguments flow through the runtime path. The callback and
// array_map paths retain normal array_merge composition.
$dynamic = dynamic_merge($first, $second, $third);
$forwarded = forward_merge($first, $third, fn($left, $right) => array_merge($left, $right));
$mapped = array_map('merge_callback', [['one' => 1, -1 => 'minus'], ['two' => 2]]);

echo 'counts=' . count($zero) . ':' . count(array_merge($empty, [])) . ':' . count($one) . "\n";
$ignored = dump_entries('zero', $zero);
$ignored = dump_entries('one', $one);
$ignored = dump_entries('one-input', $one_source);
echo 'one-nested=' . $one['nested']['inside'] . ':' . $one_source['nested']['inside'] . "\n";
$ignored = dump_entries('two', $two);
$ignored = dump_entries('merged', $merged);
$ignored = dump_entries('first-input', $first);
$ignored = dump_entries('second-input', $second);
$ignored = dump_entries('third-input', $third);
echo 'nested=' . $merged['nested']['from'] . ':' . $third['nested']['from'] . ':' . count($merged['nested']) . ':' . count($third['nested']) . "\n";
$ignored = dump_entries('marked', $marked);
echo "\n";
$ignored = dump_entries('dynamic', $dynamic);
$ignored = dump_entries('forwarded', $forwarded);
$ignored = dump_entries('mapped-first', $mapped[0]);
$ignored = dump_entries('mapped-second', $mapped[1]);
