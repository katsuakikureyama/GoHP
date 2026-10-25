<?php
// Owned differential fixture: capture-free arrows forwarded through named
// array/callable :bool functions. It intentionally composes the supported
// parameter and return boundaries without adding PHP callable lookup.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function value_is_yes($value): bool {
    return $value === 'yes';
}

function branch_match(array $items, callable $callback, $mode): bool {
    if ($mode === 'first') {
        return $callback($items[0]);
    } elseif ($mode === 'second') {
        return $callback($items[1]);
    } else if ($mode === 'scan') {
        foreach ($items as $value) {
            if ($callback($value)) {
                return true;
            }
        }
        return false;
    } else {
        return $callback(false);
    }
}

function forward_match(array $items, callable $callback, $mode): bool {
    return branch_match($items, $callback, $mode);
}

function local_write_append(array $items): bool {
    $items['changed'] = 'yes';
    $items[] = 'tail';
    return isset($items['changed']) && count($items) === 3;
}

function callback_marker($value): bool {
    echo '<callback:' . $value . '>';
    return $value === 'yes';
}

$yes = fn($value) => value_is_yes($value);
$first = forward_match(['yes', 'no'], $yes, 'first');
$second = forward_match(['no', 'yes'], $yes, 'second');
$scan = forward_match(['no', 'yes'], $yes, 'scan');
$fallback = forward_match(['yes'], $yes, 'other');
$cast = forward_match([''], fn($value) => (bool) $value, 'first');
$empty = forward_match([], $yes, 'scan');

$marked = fn($value) => callback_marker($value);
$marker_first = forward_match(['yes', 'no'], $marked, 'first');
$marker_empty = forward_match([], $marked, 'scan');

$source = ['seed'];
$mutator = fn($items) => local_write_append($items);
$local_change = $mutator($source);

echo 'paths:' . bool_word($first) . ':' . bool_word($second) . ':' . bool_word($scan) . ':' . bool_word($fallback) . "\n";
echo 'cast-empty:' . bool_word($cast) . "\n";
echo 'empty:' . bool_word($empty) . "\n";
echo 'marker:' . bool_word($marker_first) . ':' . bool_word($marker_empty) . "\n";
echo 'source:' . bool_word($local_change) . ':' . count($source) . ':' . (isset($source['changed']) ? 'changed' : 'clean') . "\n";
