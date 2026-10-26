<?php
// Owned differential fixture for the scalar-only local-frame path. It
// compares ordinary PHP values only; it does not claim PHP reference support.

function copied_locals() {
    $source = 7;
    $copy = $source;
    $copy = 9;
    return $source . ':' . $copy;
}

function scalar_edges() {
    $text = "A\tB";
    $utf8 = '雪';
    $empty = '';
    $false_value = false;
    $null_value = null;
    $lowest = -9223372036854775807;
    $highest = 9223372036854775807;
    $logic = ($false_value === false && $null_value === null && $empty !== $utf8) || false;
    $present = isset($false_value) && !isset($null_value) && !isset($false_value, $null_value);
    $bounds = ($lowest + 1) . ':' . ($highest - 1);
    return $text . ':' . $utf8 . ':' . $bounds . ':'
        . ($logic ? 'logic' : 'wrong') . ':' . ($present ? 'present' : 'wrong') . ':'
        . ((bool) $empty ? 'T' : 'F') . ':' . ((bool) $false_value ? 'T' : 'F') . ':'
        . ((bool) $null_value ? 'T' : 'F');
}

function loop_trace() {
    $index = 0;
    $trace = '';
    while ($index !== 5) {
        $index = $index + 1;
        if ($index === 2) {
            $trace = $trace . 'C';
            continue;
        }
        if ($index === 4) {
            $trace = $trace . 'B';
            break;
        }
        $trace = $trace . $index;
    }
    return $trace . ':' . $index;
}

function early_path() {
    $value = 0;
    if ($value === 0) {
        return 'early';
    }
    return 'wrong';
}

function fallthrough_path() {
    $value = 'fall';
    if (false) {
        return 'wrong';
    }
    $value = $value . 'through';
    echo $value;
}

function bare_return_path() {
    if (true) {
        return;
    }
    return 'wrong';
}

function repeated_local() {
    $local = 0;
    $local = $local + 1;
    return $local;
}

// This untyped array parameter deliberately remains on the ordinary map path.
function append_copy($items) {
    $copy = $items;
    $copy[] = 'tail';
    return $copy;
}

$source = ['head'];
$appended = append_copy($source);
$arrow = fn($value) => $value . ':arrow';

echo 'copy=' . copied_locals() . "\n";
echo 'edges=' . scalar_edges() . "\n";
echo 'loop=' . loop_trace() . "\n";
echo 'returns=' . early_path() . ':' . (fallthrough_path() === null ? 'implicit-null' : 'wrong') . ':'
    . (bare_return_path() === null ? 'null' : 'wrong') . "\n";
echo 'repeat=' . repeated_local() . ':' . repeated_local() . ':' . repeated_local() . "\n";
echo 'array=' . $source[0] . ':' . $appended[0] . ':' . $appended[1] . "\n";
echo 'arrow=' . $arrow('base') . "\n";
