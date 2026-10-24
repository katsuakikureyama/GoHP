<?php
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function has_key($key, $value) {
    return array_key_exists($key, $value);
}

$modern = ['null' => null, 'false' => false, 'zero' => 0, 'empty' => '', '2' => 'two', '02' => 'leading', '+2' => 'plus', 'nested' => ['inner' => 'inside']];
$legacy = array('null' => null, 'false' => false, '2' => 'legacy-two', '02' => 'legacy-leading', '+2' => 'legacy-plus');
$empty = array();
$integer_only = [2 => 'int-two'];
$leading_only = ['02' => 'leading-only'];

echo 'empty=' . bool_word(has_key('missing', $empty)) . "\n";
echo 'values=' . bool_word(ArRaY_KeY_ExIsTs('null', $modern)) . ':' . bool_word(has_key('false', $modern)) . ':' . bool_word(has_key('zero', $modern)) . ':' . bool_word(has_key('empty', $modern)) . "\n";
echo 'absent-present=' . bool_word(has_key('missing', $modern)) . ':' . bool_word(has_key('null', $modern)) . "\n";
echo 'numeric=' . bool_word(has_key('2', $modern)) . ':' . bool_word(has_key(2, $modern)) . ':' . bool_word(has_key('02', $modern)) . ':' . bool_word(has_key('+2', $modern)) . "\n";
echo 'numeric-absent=' . bool_word(has_key('02', $integer_only)) . ':' . bool_word(has_key('+2', $integer_only)) . ':' . bool_word(has_key(2, $leading_only)) . "\n";
echo 'nested=' . bool_word(has_key('nested', $modern)) . ':' . bool_word(has_key('inner', $modern)) . ':' . bool_word(has_key('inner', $modern['nested'])) . "\n";
echo 'legacy=' . bool_word(has_key('null', $legacy)) . ':' . bool_word(has_key(2, $legacy)) . ':' . bool_word(has_key('02', $legacy)) . ':' . bool_word(has_key('+2', $legacy)) . "\n";
if (has_key('missing', $modern) && $modern['missing'] === 'unreachable') {
    echo 'bad';
}
echo 'guarded=ok' . "\n";
echo 'original=' . $modern[2] . ':' . $modern['02'] . ':' . $modern['+2'] . ':' . $modern['nested']['inner'] . "\n";
