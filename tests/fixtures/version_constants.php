<?php
// Differential fixture for the two selected PHP targets. The predefined
// constants are ordinary integer expressions; their values differ by target.

function PHP_MINOR_VERSION() { return 90; }

function version_identity() {
    return PHP_MAJOR_VERSION === 8 ? PHP_MINOR_VERSION : 0;
}

function typed_increment(int $value) {
    return $value + 1;
}

class VersionBox {
    public function describe() {
        if (PHP_MINOR_VERSION === 3) { return 'minor-three'; }
        return 'minor-four';
    }
}

$values = [PHP_MINOR_VERSION, PHP_MAJOR_VERSION + PHP_MINOR_VERSION];
$key = 0;
$arrow = fn($value) => $value + PHP_MINOR_VERSION;
$box = new VersionBox();

echo 'identity=' . version_identity() . "\n";
echo 'arithmetic=' . typed_increment(PHP_MINOR_VERSION) . ':' . (PHP_MAJOR_VERSION + PHP_MINOR_VERSION) . "\n";
echo 'branch=' . $box->describe() . "\n";
echo 'array=' . $values[$key] . ':' . $values[1] . "\n";
echo 'reuse=' . $arrow(PHP_MINOR_VERSION) . ':' . $box->describe() . "\n";
echo 'names=' . PHP_MINOR_VERSION . ':' . PHP_MINOR_VERSION() . "\n";
$by_minor = [3 => 'three', 4 => 'four'];
$minor_key = PHP_MINOR_VERSION;
$by_minor[$minor_key] = $by_minor[PHP_MINOR_VERSION] . '-selected';
$PHP_MINOR_VERSION = 99;
echo 'keys=' . $by_minor[PHP_MINOR_VERSION] . ':' . $PHP_MINOR_VERSION . ':' . PHP_MINOR_VERSION . "\n";
