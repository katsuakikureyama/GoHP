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

function list_for($value) {
    return array_is_list($value);
}

$empty_short = [];
$empty_legacy = array();
$implicit = [null, false];
$explicit = [0 => 0, 1 => ''];
$numeric_strings = ['0' => 'zero', '1' => 'one'];
$double_zero = ['00' => 'double-zero'];
$plus_zero = ['+0' => 'plus-zero'];
$out_of_order = [1 => 'one', 0 => 'zero'];
$gap = [0 => 'zero', 2 => 'two'];
$start = [1 => 'one'];
$mixed = [0 => 'zero', 'name' => 'value'];
$nested = array(array('name' => 'first'), array('name' => 'second'));
$duplicate_in_order = [0 => 'before', 0 => 'after', 1 => 'one'];
$duplicate_out_of_order = [1 => 'before', 0 => 'zero', 1 => 'after'];

echo 'empty=' . bool_word(list_for($empty_short)) . ':' . bool_word(list_for($empty_legacy)) . "\n";
echo 'contiguous=' . bool_word(ArRaY_Is_LiSt($implicit)) . ':' . bool_word(list_for($explicit)) . ':' . bool_word(list_for($numeric_strings)) . "\n";
echo 'noncanonical=' . bool_word(list_for($double_zero)) . ':' . bool_word(list_for($plus_zero)) . "\n";
echo 'order=' . bool_word(list_for($out_of_order)) . ':' . bool_word(list_for($gap)) . ':' . bool_word(list_for($start)) . ':' . bool_word(list_for($mixed)) . "\n";
echo 'nested=' . bool_word(list_for($nested)) . ':' . $nested[0]['name'] . ':' . $nested[1]['name'] . "\n";
echo 'duplicates=' . bool_word(list_for($duplicate_in_order)) . ':' . bool_word(list_for($duplicate_out_of_order)) . ':' . $duplicate_in_order[0] . ':' . $duplicate_out_of_order[1] . "\n";
echo 'derived=' . bool_word(list_for(array_values($out_of_order))) . ':' . bool_word(list_for(array_keys($out_of_order))) . "\n";
echo 'original=' . $explicit[0] . ':' . $explicit[1] . ':' . $out_of_order[1] . ':' . $out_of_order[0] . "\n";
