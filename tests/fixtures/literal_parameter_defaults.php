<?php
// Owned differential fixture: literal parameter defaults beyond null.
// Covers int, bool, string, null, and empty-array defaults on functions,
// public methods, constructors, and public static methods.
function int_default($x = 10) {
    echo "int=" . $x . "\n";
}
function bool_false_default($x = false) {
    echo "bool=" . gettype($x) . ":" . ($x === false ? 'false' : 'true') . "\n";
}
function bool_true_default($x = true) {
    echo "bool=" . gettype($x) . ":" . ($x === true ? 'true' : 'false') . "\n";
}
function string_empty_default($x = '') {
    echo "str_empty=" . gettype($x) . ":" . strlen($x) . "\n";
}
function string_text_default($x = 'text') {
    echo "str_text=" . $x . "\n";
}
function string_dq_default($x = "dq") {
    echo "str_dq=" . $x . "\n";
}
function null_default($x = null) {
    echo "null=" . gettype($x) . "\n";
}
function empty_array_default($items = []) {
    $items[] = 1;
    echo "arr=" . count($items) . "\n";
}
function multi_defaults($a, $b = 1, $c = 'x', $d = false) {
    echo "multi=" . $a . ":" . $b . ":" . $c . ":" . ($d === false ? 'F' : 'T') . "\n";
}
class DefaultsClass {
    public $val;
    public function method($x = 42) {
        echo "method=" . $x . "\n";
    }
    public function __construct($a = 'constructed', $b = null) {
        $this->val = $a;
    }
    public static function staticMethod($y = true) {
        echo "static=" . ($y === true ? 'yes' : 'no') . "\n";
    }
}
// Int default: omitted vs passed
int_default();
int_default(20);
// Bool defaults
bool_false_default();
bool_false_default(true);
bool_true_default();
bool_true_default(false);
// String defaults
string_empty_default();
string_text_default();
string_dq_default();
// Null default
null_default();
null_default('not null');
// Empty array default: no sharing between calls
empty_array_default();
empty_array_default();
empty_array_default([1, 2, 3]);
// Multiple trailing defaults with only some passed
multi_defaults('A');
multi_defaults('B', 2);
multi_defaults('C', 3, 'Y', true);
// Public method with defaults
$obj = new DefaultsClass();
$obj->method();
$obj->method(99);
// Constructor with defaults
$obj2 = new DefaultsClass();
echo "ctor=" . $obj2->val . "\n";
$obj3 = new DefaultsClass('custom');
echo "ctor=" . $obj3->val . "\n";
// Public static method with defaults
DefaultsClass::staticMethod();
DefaultsClass::staticMethod(false);
