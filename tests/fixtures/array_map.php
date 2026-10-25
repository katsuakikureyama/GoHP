<?php
$items = [9 => "O'Reilly", 'name' => "A\\B", 3 => 'end'];
$mapped = ArRaY_MaP('mapped_escape', $items);
foreach ($mapped as $key => $value) {
    echo $key . ':' . $value . ',';
}
echo "\n";
echo count(array_map('mapped_escape', [])) . "\n";
echo $items[9] . ':' . $items['name'] . "\n";

function mapped_escape($value) {
    echo '<' . $value . '>';
    return AdDsLaShEs($value);
}

function recursive_escape($value) {
    if (is_array($value)) {
        return array_map('recursive_escape', $value);
    }
    if (is_string($value)) {
        return addslashes($value);
    }
    return $value;
}

$nested = ['branch' => [4 => "D'Angelo", 'quote' => '"'], 'flag' => false, 'number' => 42, 'nothing' => null];
$escaped = array_map('recursive_escape', $nested);
echo $escaped['branch'][4] . ':' . $escaped['branch']['quote'] . "\n";
echo $nested['branch'][4] . ':' . $nested['branch']['quote'] . "\n";
echo is_bool($escaped['flag']) . ':' . ($escaped['number'] === 42) . ':' . ($escaped['nothing'] === null) . "\n";
echo addslashes(stripslashes('A\\0B')) . "\n";
