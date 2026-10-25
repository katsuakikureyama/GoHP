<?php
function keys_for($value) {
    return array_keys($value);
}
function returned_array() {
    return ['returned' => 1, 7 => 'seven'];
}
$empty = [];
$items = ['first' => 'one', 9 => 'nine', '2' => 'before', '02' => 'leading', 2 => 'replacement', 'nested' => ['inside' => 1]];
$keys = array_keys($items);
echo count(array_keys($empty)) . "\n";
echo $keys[0] . ':' . $keys[1] . ':' . $keys[2] . ':' . $keys[3] . ':' . $keys[4] . "\n";
foreach ($keys as $index => $key) {
    echo $index . ':' . $key . ',';
}
echo "\n";
echo $items[2] . ':' . $items['02'] . ':' . count($items) . "\n";
foreach (ARRAY_KEYS(returned_array()) as $key) {
    echo $key . ',';
}
echo "\n";
foreach (keys_for($items) as $key) {
    echo $key . ',';
}
echo "\n";
foreach (array_keys(array_keys($items)) as $key) {
    echo $key . ',';
}
echo "\n";
