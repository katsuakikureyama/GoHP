<?php
// A real program using a function, mixed ordered arrays, a class and native method dispatch.
function total($values) {
    $sum = 0;
    foreach ($values as $value) {
        $sum = $sum + $value;
    }
    return $sum;
}
class Receipt {
    public $prefix = 'total=';
    public function render($items) {
        return $this->prefix . total($items);
    }
}
$values = [2, 'bonus' => 5, 3];
$receipt = new Receipt();
echo $receipt->render($values) . "\n";
foreach ($values as $key => $value) {
    echo $key . ':' . $value . "\n";
}
echo count($values) . "\n";
