<?php
require __DIR__ . '/lib/dep.php';

$greeting = dep_greet('world');
$sum = add(3, 4);
$obj = new HoistedBox('item');
$label = $obj->tag;
$len = strlen($label);

echo 'greet=' . $greeting . "\n";
echo 'sum=' . $sum . "\n";
echo 'label=' . $label . "\n";
echo 'len=' . $len . "\n";

function add($a, $b) {
    return $a + $b;
}

final class HoistedBox {
    public $tag;

    public function __construct($tag) {
        $this->tag = $tag;
    }
}
