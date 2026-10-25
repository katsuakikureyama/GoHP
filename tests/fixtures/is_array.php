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

function dynamic_value($value) {
    return $value;
}

function marked_value($value) {
    echo '!';
    return $value;
}

function guarded_count($value) {
    if (is_array($value) && count($value) === 2) {
        return 'two';
    }
    return 'other';
}

function guarded_list($value) {
    if (is_array($value) && array_is_list($value)) {
        return 'list';
    }
    return 'other';
}

class ArrayLike {
    public $contents = 'not-a-php-array';
}

$empty_short = [];
$empty_legacy = array();
$list = ['first', 'second'];
$assoc = ['first' => 'one', 'second' => 'two'];
$nested = array(array('inner' => 'value'));
$object = new ArrayLike();

echo 'arrays=' . bool_word(Is_ArRaY($empty_short)) . ':' . bool_word(is_array($empty_legacy)) . ':' . bool_word(is_array($list)) . ':' . bool_word(is_array($assoc)) . ':' . bool_word(is_array($nested)) . "\n";
echo 'scalars=' . bool_word(is_array(null)) . ':' . bool_word(is_array(false)) . ':' . bool_word(is_array(true)) . ':' . bool_word(is_array(0)) . ':' . bool_word(is_array('array')) . ':' . bool_word(is_array('0')) . ':' . bool_word(is_array($object)) . "\n";
echo 'dynamic=' . bool_word(is_array(dynamic_value($list))) . ':' . bool_word(is_array(dynamic_value($object))) . "\n";
echo 'marker=' . bool_word(is_array(marked_value($assoc))) . "\n";
echo 'guards=' . guarded_count($list) . ':' . guarded_count($object) . ':' . guarded_list($list) . ':' . guarded_list($object) . "\n";
echo 'original=' . $list[0] . ':' . $list[1] . ':' . $assoc['first'] . ':' . $nested[0]['inner'] . "\n";
