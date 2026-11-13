<?php
// Differential fixture: a final class behaves like an ordinary class in the
// subset because inheritance is unsupported.
final class Marker {
    public $label;

    public function __construct($label) {
        $this->label = $label;
    }

    public function describe() {
        return "marker:" . $this->label;
    }
}

$first = new Marker("a");
$second = new Marker("b");
echo $first->describe() . "\n";
echo $second->describe() . "\n";
echo ($first === $second ? "same" : "distinct") . "\n";
echo ($first === $first ? "same" : "distinct") . "\n";
