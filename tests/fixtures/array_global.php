<?php
// Array- and object-valued source globals (step 6) differential fixture.
// Tests: registry pattern, counter arrays, append, isset/count/empty,
// copy semantics, cross-function writes, evaluation order.

final class Box {
    public $items;
    public function __construct() { $this->items = array(); }
    public function add($v) { $this->items[] = $v; }
    public function count() { return count($this->items); }
}

$reg = array();
$counts = array();
$stack = array();

function add_hook($name, $value) {
    global $reg;
    if (!isset($reg[$name])) {
        $reg[$name] = new Box();
    }
    $reg[$name]->add($value);
}

function increment($name) {
    global $counts;
    if (!isset($counts[$name])) {
        $counts[$name] = 1;
    } else {
        $counts[$name] = $counts[$name] + 1;
    }
}

function push_item($x) {
    global $stack;
    $stack[] = $x;
}

function reg_count() {
    global $reg;
    return count($reg);
}

function reg_empty() {
    global $reg;
    return empty($reg);
}

function reg_isset() {
    global $reg;
    return isset($reg);
}

function reg_truthy() {
    global $reg;
    if ($reg) {
        return "yes";
    }
    return "no";
}

function copy_test() {
    global $counts;
    $copy = $counts;
    $copy["a"] = 99;
    return $counts["a"];
}

function writer($key, $val) {
    global $counts;
    $counts[$key] = $val;
}

function reader($key) {
    global $counts;
    return $counts[$key];
}

function order_value() {
    global $order;
    $order[] = "inner";
    return "outer";
}

function order_test() {
    global $order;
    $order = array();
    $order["k"] = order_value();
    return count($order) . ":" . $order["k"] . ":" . $order[0];
}

// Registry pattern
add_hook("init", "hello");
add_hook("init", "world");
add_hook("other", "foo");
echo "hooks:" . $reg["init"]->count() . ":" . $reg["other"]->count() . "\n";

// Counter array
increment("a");
increment("a");
increment("b");
echo "counts:" . $counts["a"] . ":" . $counts["b"] . "\n";

// Stack append
push_item(1);
push_item(2);
push_item(3);
echo "stack:" . count($stack) . "\n";

// count/empty/isset/truthy on globals
echo "reg_count:" . reg_count() . "\n";
echo "reg_empty:" . (reg_empty() ? "true" : "false") . "\n";
echo "reg_isset:" . (reg_isset() ? "true" : "false") . "\n";
echo "reg_truthy:" . reg_truthy() . "\n";

// Copy semantics: $copy = $counts; $copy["a"] = 99; global unchanged
echo "copy:" . copy_test() . "\n";

// Cross-function write/read
writer("x", 42);
echo "cross:" . reader("x") . "\n";

// Evaluation order: $order["k"] = f() where f appends to the same global.
// PHP evaluates the right-hand side before fetching the container, so f's
// append survives.
echo "order:" . order_test() . "\n";
