<?php
declare(strict_types=1);

// Differential fixture for public static methods, Class::method() calls,
// final classes and the strict_types declaration no-op.
final class Calculator {
    public $value;

    public function __construct($value) {
        $this->value = $value;
    }

    public static function add(int $a, int $b) {
        return $a + $b;
    }

    public static function twice(int $value) {
        return Calculator::add($value, $value);
    }

    public static function make($value) {
        return new Calculator($value);
    }

    public function plus($delta) {
        return Calculator::add($this->value, $delta);
    }

    public static function sumArray($values) {
        $total = 0;
        foreach ($values as $value) {
            $total = $total + $value;
        }
        return $total;
    }

    public static function mutate($values) {
        $values[0] = 999;
        return $values[0];
    }
}

function globalAdd(int $a, int $b) {
    return Calculator::add($a, $b);
}

echo 'add:' . Calculator::add(2, 3) . "\n";
echo 'twice:' . Calculator::twice(4) . "\n";
echo 'case:' . cAlCuLaToR::AdD(5, 6) . "\n";
echo 'global:' . globalAdd(7, 8) . "\n";

$made = Calculator::make(10);
echo 'factory:' . $made->plus(5) . "\n";

$values = [1, 2, 3];
echo 'sum:' . Calculator::sumArray($values) . "\n";
echo 'copy:' . Calculator::mutate($values) . ':' . $values[0] . "\n";

$direct = new Calculator(20);
echo 'instance-method:' . $direct->plus(6) . "\n";
