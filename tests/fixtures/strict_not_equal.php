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

$values = [0, '1', null, false];
$minimum = -9223372036854775807 + -1;
$state = 'cold';

echo 'same=' . bool_check((null !== null)) . ',' . bool_check((false !== false)) . ',' . bool_check((true !== false)) . ',' . bool_check((0 !== 0)) . ',' . bool_check((-7 !== -8)) . ',' . bool_check(('same' !== 'same')) . ',' . bool_check(('left' !== 'right')) . "\n";
echo 'cross=' . bool_check((1 !== '1')) . ',' . bool_check((null !== false)) . ',' . bool_check((0 !== false)) . ',' . bool_check(('' !== null)) . ',' . bool_check(('0' !== 0)) . "\n";
echo 'cross-more=' . bool_check(0 !== null) . ',' . bool_check('' !== false) . "\n";
echo 'integers=' . bool_check(($minimum !== -9223372036854775807)) . ',' . bool_check(($minimum !== -9223372036854775807 + -1)) . ',' . bool_check((-9223372036854775807 !== -9223372036854775807)) . "\n";
echo 'maximum=' . bool_check(9223372036854775807 !== 9223372036854775807) . ',' . bool_check(9223372036854775807 !== '9223372036854775807') . "\n";
echo 'strings=' . bool_check(('01' !== '1')) . ',' . bool_check(('1' !== 1)) . ',' . bool_check(('東京' !== '東京')) . ',' . bool_check(('東京' !== '京都')) . "\n";

if ((1 !== 2)) {
    $state = 'warm';
}
echo 'if=' . $state . "\n";

echo 'once=' . bool_check((marker('left', 1) !== marker('right', 2))) . "\n";
echo 'short=' . bool_check(false && (marker('and-skip', 1) !== 2)) . ',' . bool_check(true || (marker('or-skip', 1) !== 2)) . "\n";
echo 'precedence=' . bool_check((!false !== true)) . ',' . bool_check((1 + 2 !== 3)) . ',' . bool_check(('a' . 'b' !== 'ab')) . ',' . bool_check(true && (1 !== 2)) . ',' . bool_check(false || (1 !== 1)) . "\n";
echo 'mixed=' . bool_check(((1 !== 2) === true)) . ',' . bool_check(((1 !== 1) !== false)) . "\n";
echo 'binding=' . bool_check(false && 1 !== 2) . ',' . bool_check(true || false !== true) . ',' . bool_check(true === (1 !== 2)) . ',' . bool_check(1 !== (2 === true)) . "\n";
echo 'index=' . bool_check(($values[0] !== $values[1])) . ',' . bool_check(($values[2] !== $values[2])) . ',' . bool_check(($values[3] !== $values[0])) . "\n";
echo 'ternary=' . ((marker('ternary-left', 1) !== marker('ternary-right', 1)) ? marker('never', 'N') : marker('chosen', 'Y')) . "\n";
echo 'ternary-binding=' . (1 !== 1 ? 'bad' : 'ok') . ':' . (1 !== '1' ? 'ok' : 'bad') . "\n";
