<?php
function rewrite_tree($items) {
    foreach ($items as $key => $value) {
        if (is_array($value)) {
            $result = rewrite_tree($value);
            $items[$key] = $result;
        }
        if (is_string($value)) {
            $items[$key] = $value . '!';
        }
    }
    return $items;
}

function marker() {
    echo 'M';
    return 'rhs';
}

$source = [3 => 'three', '2' => 'two', 'branch' => [7 => 'deep', 'leaf' => 'leaf']];
$copy = $source;
$written = rewrite_tree($copy);
echo $source[3] . ':' . $source[2] . ':' . $source['branch'][7] . "\n";
echo $written[3] . ':' . $written[2] . ':' . $written['branch'][7] . ':' . $written['branch']['leaf'] . "\n";
$copy[3] = 'copy-only';
echo 'copy=' . $source[3] . ':' . $copy[3] . ':' . $written[3] . "\n";

$normal = ['2' => 'before', 'tail' => 'tail'];
$key = '2';
$normal[$key] = 'replaced';
$normal['later'] = 'inserted';
$normal['02'] = 'distinct';
foreach ($normal as $entry_key => $entry_value) {
    echo $entry_key . ':' . $entry_value . ',';
}
echo "\n";

$same = [0 => 'old'];
$same[0] = $same[0] . '!';
echo $same[0] . "\n";

$one = ['value' => 'before'];
$one['value'] = marker();
echo $one['value'] . "\n";

$holder = [];
$inner = ['value' => 'base'];
$holder['child'] = $inner;
$inner['value'] = 'changed';
echo $holder['child']['value'] . ':' . $inner['value'] . "\n";

$iterated = '';
foreach ($written as $key => $value) {
    $iterated = $iterated . $key . ',';
    $written['late'] = 'new entry';
}
echo $iterated . ':' . count($written) . ':' . $written['late'] . "\n";
