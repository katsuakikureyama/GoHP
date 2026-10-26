<?php
// Differential coverage for the first generated ValueSlot lane. These are
// nullary, straight-line functions with scalar and flat-array locals only.

function owned_copy_isolation() {
    $source = [-5 => 'before', 'label' => null];
    $copy = $source;
    $copy['label'] = 'copy';
    $copy[] = 'auto';
    $source[-5] = 'after';
    $source = ['replaced' => 'source'];
    return $copy[-5] . ':' . $copy['label'] . ':' . $copy[-4] . ':' . count($copy) . ':' . count($source);
}

function owned_self_assignment() {
    $items = ['first' => 'one'];
    $items = $items;
    $items[] = 'two';
    return $items['first'] . ':' . $items[0] . ':' . count($items);
}

function owned_keys_and_null() {
    $items = [-5 => 'negative', 'label' => null];
    $items[] = 'auto';
    $items['tail'] = 'end';
    return (isset($items['label']) ? 'set' : 'unset') . ':' . $items[-5] . ':' . $items[-4] . ':' . $items['tail'] . ':' . count($items);
}

function owned_ternary_snapshot() {
    $before = ['value' => 'before'];
    $after = ['value' => 'after'];
    $pick = true;
    $selected = $pick ? $before : $after;
    $before['value'] = 'changed';
    return $selected['value'];
}

function owned_ternary_missing() {
    $marker = [];
    $pick = true;
    $safe = 'selected';
    $missing = ['known' => 'never'];
    return $pick ? $safe : $missing['missing'];
}

function owned_ternary_overflow() {
    $marker = [];
    $pick = true;
    $safe = 'selected';
    $edge = 9223372036854775807;
    return $pick ? $safe : $edge + 1;
}

function owned_mixed_locals() {
    $items = ['key' => 'value'];
    $integer = 3;
    $word = 'go';
    $flag = false;
    $nothing = null;
    $items['number'] = $integer + 2;
    $flag = $nothing === null && $items['number'] === 5;
    $lazy = true || $items['missing'];
    $skip = false && $items['missing'];
    return ($flag && $lazy && !$skip ? $word : 'bad') . ':' . $items['key'] . ':' . $items['number'];
}

function owned_returned_array() {
    $items = [-5 => 'negative', 'label' => null];
    $items[] = 'auto';
    return $items;
}

// This is intentionally not invoked by the positive fixture. The generated
// ownership lane must still preserve the existing typed-array return guard.
function owned_bad_return(): array {
    $items = ['value' => null];
    return $items['value'];
}

// Only the external consumer invokes this unsupported-overflow path.
function owned_overflow() {
    $items = ['edge' => 9223372036854775807];
    return $items['edge'] + 1;
}

// This stays on the legacy lane: the untyped host value can be nested.
function legacy_identity($value) {
    return $value;
}

// These intentionally keep the legacy lane and share an Execution in main.
function legacy_state_start() {
    return strtok('first,second', ',');
}

function legacy_state_next() {
    return strtok(',');
}

echo 'copy=' . owned_copy_isolation() . "\n";
echo 'self=' . owned_self_assignment() . "\n";
echo 'keys=' . owned_keys_and_null() . "\n";
echo 'ternary=' . owned_ternary_snapshot() . ':' . owned_ternary_missing() . ':' . owned_ternary_overflow() . "\n";
echo 'mixed=' . owned_mixed_locals() . "\n";
$first = owned_returned_array();
$second = owned_returned_array();
$first['label'] = 'first';
echo 'returned=' . $first[-5] . ':' . $first[-4] . ':' . $first['label'] . ':' . (isset($second['label']) ? 'set' : 'null') . "\n";
foreach ($second as $key => $value) {
    echo gettype($key) . ':' . $key . '=' . gettype($value) . ':' . $value . ';';
}
echo "\n";
$legacy = ['key' => 'legacy'];
foreach ($legacy as $key => $value) {
    echo 'legacy=' . $key . ':' . $value . "\n";
}
echo 'state=' . legacy_state_start() . ':' . legacy_state_next() . "\n";
