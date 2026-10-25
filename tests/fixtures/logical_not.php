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

function marker($value) {
    echo $value;
    return '';
}

$empty = [];
$items = [0, 'value'];
$zero = 0;

echo 'values=' . bool_check(!null) . ',' . bool_check(!true) . ',' . bool_check(!false) . ',' . bool_check(!0) . ',' . bool_check(!-7) . ',' . bool_check(!'') . ',' . bool_check(!'0') . ',' . bool_check(!'00') . ',' . bool_check(!'value') . ',' . bool_check(!$empty) . ',' . bool_check(!$items) . "\n";
echo 'double=' . bool_word(!!null) . ':' . bool_word(!!'0') . ':' . bool_word(!!'00') . ':' . bool_word(!!$items) . "\n";

$state = 'full';
if (!$empty) {
    $state = 'empty';
}
echo 'if=' . $state;
echo ',ternary=' . (!$items ? 'empty' : 'full') . ':' . (!$zero ? 'zero' : 'nonzero') . "\n";

echo 'short=' . (!true && marker('never')) . ':' . (!false && marker('once')) . "\n";
echo 'precedence=' . !0 + 2 . ':' . bool_word(!(0 + 2)) . ':' . bool_word(!0 === true) . ':' . bool_word(!0 === false) . "\n";
echo 'index=' . bool_check(!$items[0]) . ',' . bool_check(!$items[1]) . "\n";
echo 'operand=' . bool_check(!marker('call')) . ':' . bool_check(!!marker('double')) . "\n";
echo 'casts=' . bool_check(!(bool) '0') . ':' . bool_check((bool) !'0') . "\n";
echo 'lazy=' . (!'0' ? marker('chosen') : marker('never-ternary')) . "\n";
