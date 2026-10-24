<?php
function copied($array) { return $array; }
$a = ['1' => 4, 1 => 7, 'name' => 'ok', 3, [8, 9]];
$b = copied($a);
echo $b[1] . ':' . $b['name'] . ':' . $b[2] . ':' . $b[3][1] . "\n";
echo count($b) . "\n";
echo true . false . null . "\n";
echo ('12' + 3) . "\n";
