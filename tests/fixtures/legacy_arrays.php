<?php
function legacy_empty() {
    return array();
}

function legacy_values() {
    return ArRaY(
        'zero',
        3 => 'three',
        'short' => ['one', 'two',],
    );
}

function short_values() {
    return [
        'legacy' => aRrAy('left', 'right',),
    ];
}

$empty = legacy_empty();
$values = legacy_values();
$short = short_values();
echo count($empty) . ':' . $values[0] . ':' . $values[3] . ':' . $values['short'][1] . ':' . $short['legacy'][0] . "\n";
foreach (array('left' => 'L', 5 => 'R',) as $key => $value) {
    echo $key . '=' . $value . ';';
}
echo "\n";
