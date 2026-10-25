<?php
function values_for($value) {
    return array_values($value);
}
function returned_array() {
    return ['returned' => 'return', 8 => 'eight'];
}
$empty = [];
$items = [9 => 'nine', 'label' => 7, 3 => true, 20 => null, 'nested' => ['inside' => 'value'], '2' => 'before', 2 => 'replacement'];
$values = ARRAY_VALUES($items);
$nested_values = array_values($values);
echo count(array_values($empty)) . "\n";
echo count($values) . "\n";
echo $values[0] . ':' . $values[1] . ':' . $values[2] . ':' . $values[3] . ':' . $values[5] . "\n";
echo $items[2] . ':' . count($items) . "\n";
echo $values[4]['inside'] . ':' . array_keys($values[4])[0] . "\n";
echo $nested_values[5] . "\n";
foreach (array_values(array_keys($values)) as $index => $value) {
    echo $index . ':' . $value . ',';
}
echo "\n";
foreach (values_for(returned_array()) as $index => $value) {
    echo $index . ':' . $value . ',';
}
echo "\n";
