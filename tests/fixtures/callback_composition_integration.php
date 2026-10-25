<?php
// Original composition fixture: a literal recursive callback preserves shape
// while selecting only the bounded transform for each runtime value kind.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function escape_tree($value) {
    echo is_array($value) ? '<A>' : (is_string($value) ? '<S>' : '<X>');
    return is_array($value) ? array_map('escape_tree', $value) : (is_string($value) ? addslashes($value) : $value);
}

function unescape_tree($value) {
    return is_array($value) ? array_map('unescape_tree', $value) : (is_string($value) ? stripslashes($value) : $value);
}

$source = array(
    'phrase' => "O'Reilly",
    7 => array('slash' => "A\\B", 'nul' => stripslashes('x\\0y'), 'empty' => array()),
    'none' => null,
    'flag' => true,
    'number' => 42,
    'tail' => 'Z',
);

echo 'calls=';
$escaped = array_map('escape_tree', $source);
echo "\n";
$empty = array_map('escape_tree', array());
$round_trip = array_map('unescape_tree', $escaped);
$keys = array_keys($round_trip);
$values = array_values($round_trip);

echo 'empty=' . count($empty) . ':' . bool_word(array_is_list($empty)) . "\n";
echo 'keys=' . $keys[0] . ':' . $keys[1] . ':' . $keys[2] . ':' . bool_word(array_is_list($keys)) . ':' . bool_word(array_is_list($values)) . "\n";
echo 'nested=' . $round_trip[7]['slash'] . ':' . strlen($round_trip[7]['nul']) . ':' . bool_word($round_trip[7]['nul'] === $source[7]['nul']) . ':' . bool_word(array_key_exists('empty', $round_trip[7])) . ':' . bool_word(array_is_list($round_trip[7]['empty'])) . "\n";
echo 'types=' . bool_word($round_trip['none'] === null) . ':' . bool_word($round_trip['flag'] === true) . ':' . bool_word($round_trip['number'] === 42) . ':' . bool_word(array_key_exists('none', $round_trip)) . "\n";
echo 'string=' . sprintf('%010s', str_replace('O', 'o', $round_trip['phrase'])) . ':' . $round_trip['tail'] . "\n";
echo 'original=' . $source['phrase'] . ':' . $source[7]['slash'] . ':' . strlen($source[7]['nul']) . ':' . $source['tail'] . "\n";
