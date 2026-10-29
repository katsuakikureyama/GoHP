<?php
// Owned differential fixture for the deliberately narrow array_push statement
// lane. It uses only a bare local or one declared public property as target.

function push_marker($value) {
    echo '<push:' . $value . '>';
    return $value;
}

function reset_box($box) {
    $box->items = ['reset'];
    return 'added';
}

function push_unknown_receiver($box, $value) {
    $count = array_push($box->items, $value);
    return $count;
}

class PushChild {
    public $name;

    public function __construct($name) {
        $this->name = $name;
    }
}

class PushBox {
    public $items;

    public function __construct() {
        $this->items = [];
    }
}

class PushFrame {
    public $name;
    public $stack;

    public function __construct($name) {
        $this->name = $name;
        $this->stack = [];
    }

    public function push_new($name) {
        $count = array_push($this->stack, new PushFrame($name));
        return $count;
    }
}

// A target with no values is still an array call and returns its integer size.
$empty = [];
$zero = array_push($empty);
$items = [];
$multiple = array_push($items, 'first', 0, -3);

// Numeric append follows the next highest integer key, including sparse keys.
$sparse = [3 => 'three', -2 => 'negative'];
$sparse_count = array_push($sparse, 'next');
$negative = [-5 => 'negative'];
$negative_count = array_push($negative, 'tail');
$string_keys = ['label' => 'label'];
$string_count = array_push($string_keys, 'zero');

// Both source copies and self payloads are snapshots, not references.
$source = ['old'];
$alias = $source;
$first_self = array_push($source, $source);
$second_self = array_push($source, $source);
$alias[0] = 'alias-only';
$together = ['same'];
$together_count = array_push($together, $together, $together);

// Array payloads are copied; a child object payload retains its identity.
$nested = ['child' => ['value' => 'source']];
$payloads = [];
array_push($payloads, $nested);
$nested['child'] = ['value' => 'source-mutated'];
$payload_copy = $payloads[0];
$payload_child = $payload_copy['child'];
$payload_child['value'] = 'payload-mutated';
$payload_copy['child'] = $payload_child;
$payloads[0] = $payload_copy;
$child = new PushChild('child-original');
array_push($payloads, $child);
$child->name = 'child-mutated';

// Direct known and untyped-receiver property targets share the same operation.
$box = new PushBox();
$known = array_push($box->items, 'known');
$unknown = push_unknown_receiver($box, 'unknown');

// Payload evaluation wins before the target is read and guarded. The first
// payload snapshots old items, while reset_box replaces the property.
$box->items = ['old'];
$reset_count = array_push($box->items, $box->items, reset_box($box));

// Even a scalar property can be repaired by its RHS before the actual guard.
$repair = new PushBox();
$repair->items = 'scalar';
$repair_count = array_push($repair->items, reset_box($repair));

// The payload calls prove left-to-right, once-only evaluation.
$ordered = [];
$ordered_count = array_push($ordered, push_marker('A'), push_marker('B'));

// A method may use $this->stack as the direct target and push a freshly made
// generated object without introducing general by-reference argument support.
$frame = new PushFrame('root');
$frame_count = $frame->push_new('child');
$pushed_frame = $frame->stack[0];
$pushed_frame->name = 'child-mutated';

echo 'basic:' . $zero . ':' . (is_int($zero) ? 'int' : 'wrong') . ':' . $multiple . ':' . $items[0] . ':' . $items[1] . ':' . $items[2] . "\n";
echo 'sparse:' . $sparse_count . ':' . $sparse[4] . "\n";
echo 'keys:' . $negative_count . ':' . $negative[-4] . ':' . $string_count . ':' . $string_keys[0] . "\n";
echo 'self:' . $first_self . ':' . $second_self . ':' . count($source[1]) . ':' . count($source[2]) . ':' . $source[1][0] . ':' . $source[2][1][0] . ':' . $alias[0] . "\n";
echo 'together:' . $together_count . ':' . count($together[1]) . ':' . count($together[2]) . ':' . $together[1][0] . ':' . $together[2][0] . "\n";
echo 'copies:' . $nested['child']['value'] . ':' . $payloads[0]['child']['value'] . ':' . $payloads[1]->name . "\n";
echo 'properties:' . $known . ':' . $unknown . ':' . $box->items[0] . ':' . $box->items[1][0] . ':' . $box->items[2] . ':' . $reset_count . "\n";
echo 'repair:' . $repair_count . ':' . $repair->items[0] . ':' . $repair->items[1] . "\n";
echo 'order:' . $ordered_count . ':' . $ordered[0] . ':' . $ordered[1] . "\n";
echo 'method:' . $frame_count . ':' . $pushed_frame->name . "\n";
