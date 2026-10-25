<?php
function bool_word($value) {
    if ($value === true) {
        return 'T';
    }
    if ($value === false) {
        return 'F';
    }
    return 'X';
}

function bool_check($value) {
    return bool_word(is_bool($value)) . ':' . bool_word($value === true) . ':' . bool_word($value === false);
}

function marker($name, $value) {
    echo '<' . $name . '>';
    return $value;
}

function guarded_count($value) {
    if (!is_array($value) || count($value) === 2) {
        return 'accept';
    }
    return 'reject';
}

$empty = [];
$items = [0, 'value'];
$state = 'cold';

echo 'false-left=' . bool_check(null || null) . ',' . bool_check(false || true) . ',' . bool_check(0 || -7) . ',' . bool_check('' || '00') . ',' . bool_check('0' || '0') . ',' . bool_check($empty || $items) . "\n";
echo 'true-left=' . bool_check(true || null) . ',' . bool_check(-7 || false) . ',' . bool_check('value' || 0) . ',' . bool_check($items || '') . ',' . bool_check(true || $empty) . "\n";
echo 'false-right=' . bool_check(false || false) . ',' . bool_check(false || 0) . ',' . bool_check(false || '') . ',' . bool_check(false || $empty) . "\n";
echo 'not-precedence=' . bool_check(!true || true) . ':' . bool_check(!(true || true)) . "\n";

if (false || true) {
    $state = 'warm';
}
echo 'if=' . $state . "\n";

echo 'short=' . bool_check(marker('a', false) || marker('b', false) || marker('c', true) || marker('d', true)) . ':' . bool_check(marker('e', true) || marker('f', false)) . "\n";
echo 'mix=' . bool_check(false && true || true) . ',' . bool_check(false && (true || true)) . ',' . bool_check(true || false && false) . "\n";
echo 'operators=' . bool_check(!'0' || false) . ',' . bool_check((bool) '0' || true) . ',' . bool_check(0 + 0 || -7) . ',' . bool_check('' . '' || 'value') . ',' . bool_check($items[0] || $items[1]) . "\n";
echo 'ternary=' . (marker('condition', false) || marker('right', true) ? marker('yes', 'Y') : marker('no', 'N')) . ':' . (true || false ? marker('yes2', 'Y') : marker('no2', 'N')) . "\n";
echo 'false-ternary=' . (marker('left-false', false) || marker('right-false', false) ? marker('never', 'Y') : marker('chosen', 'N')) . "\n";
echo 'guard=' . guarded_count(null) . ':' . guarded_count([1]) . ':' . guarded_count([1, 2]) . "\n";
