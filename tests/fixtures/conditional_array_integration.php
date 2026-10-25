<?php
// Original composition fixture: a guarded string transform preserves arrays.
function guarded_value($value) {
    return is_string($value) ? stripslashes($value) : $value;
}

function marker($value) {
    echo '<' . $value . '>';
    return $value;
}

function choose_lazy($condition, $value) {
    return $condition ? 'chosen' : stripslashes($value);
}

$source = array('word' => 'A\\0B', 'nested' => array('first', 'second'), 'none' => null, 9 => 'tail');
$kept = guarded_value($source);
$keys = array_keys($kept);
$values = array_values($kept);
$empty = guarded_value(array());

echo 'shape=' . array_is_list($source) . ':' . array_is_list($keys) . ':' . array_is_list($values) . ':' . $keys[0] . ':' . $values[1][0] . "\n";
echo 'null=' . array_key_exists('none', $kept) . ':' . (is_string($kept['none']) ? 'string' : 'not-string') . "\n";
echo 'empty=' . array_is_list($empty) . ':' . count(array_values($empty)) . "\n";
echo 'string=' . strlen(str_replace('B', 'Z', guarded_value($source['word']))) . ':' . sprintf('%04s', 'x') . "\n";
echo 'lazy=' . choose_lazy(marker('once'), $source) . "\n";
echo 'original=' . strlen($source['word']) . ':' . $source['nested'][1] . ':' . $source[9] . "\n";
