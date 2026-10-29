<?php
// Owned differential fixture for the deliberately narrow array_pop statement
// lane. Every array_pop call has exactly one existing bare-local or direct
// named-property target, and is either standalone or the complete RHS of a
// bare local assignment.

class PopChild {
    public $name;

    public function __construct($name) {
        $this->name = $name;
    }
}

class PopBlock {
    public $name;

    public function __construct($name) {
        $this->name = $name;
    }
}

class PopFrame {
    public $name;
    public $block;
    public $stack;

    public function __construct($name) {
        $this->name = $name;
        $this->block = new PopBlock($name . '-block');
        // Array defaults stay in constructors, rather than class declarations.
        $this->stack = [];
    }

    public function push_child($name) {
        $this->stack[] = new PopFrame($name);
    }

    public function pop_child() {
        $child = array_pop($this->stack);
        return $child;
    }
}

class PopBox {
    public $items;

    public function __construct() {
        $this->items = [];
    }

    public function pop_item() {
        $item = array_pop($this->items);
        return $item;
    }
}

function pop_parameter($items) {
    $item = array_pop($items);
    return $item;
}

function entries($label, $items) {
    echo $label . '=';
    foreach ($items as $key => $value) {
        echo gettype($key) . ':' . $key . ':' . gettype($value) . ':' . $value . ';';
    }
    echo "\n";
}

// Pop follows insertion order rather than numeric-key order. A lower key that
// was inserted later is removed, and that removal must not rewind a higher
// existing integer append position.
$ordered = [5 => 'five', 'label' => 'label', -2 => 'minus'];
$ordered[-9] = 'later-low';
$ordered_value = array_pop($ordered);
$ordered[] = 'tail';
$negative = [-5 => 'negative'];
$negative_value = array_pop($negative);
$negative[] = 'negative-readded';
$sparse = [0 => 'zero', 2 => 'two'];
$sparse_value = array_pop($sparse);
$sparse[] = 'two-readded';
$discard = ['discard'];
array_pop($discard);

// Numeric-string keys normalize in PHP, whereas a leading-zero spelling stays
// a string key. The ordinary pop result is assigned before inspection.
$keys = ['2' => 'two', '02' => 'leading-zero', 'name' => 'named'];
$key_value = array_pop($keys);

// Values preserve their falsy distinctions. Empty-array pop returns null.
$scalars = [0, false, '', null];
$null_value = array_pop($scalars);
$blank_value = array_pop($scalars);
$false_value = array_pop($scalars);
$zero_value = array_pop($scalars);
$nothing = [];
$empty_value = array_pop($nothing);

// An array value returned from pop is its own eager copy. A child object in an
// array retains identity when it is returned.
$source = [['nested' => ['value' => 'source']]];
$source_copy = $source;
$nested_value = array_pop($source);
$nested = $nested_value['nested'];
$nested['value'] = 'returned-mutated';
$nested_value['nested'] = $nested;
$child = new PopChild('child-original');
$object_values = [$child];
$returned_child = array_pop($object_values);
$returned_child->name = 'child-mutated';

// A by-value local parameter has its own array value; the caller's array is
// unaffected by a pop inside the function.
$argument_values = ['caller-first', 'caller-last'];
$parameter_value = pop_parameter($argument_values);

// A direct property target is shared through the object and supports both an
// assigned result and the standalone statement form.
$box = new PopBox();
$box->items = ['property-first', 'property-last'];
$property_value = $box->pop_item();
array_pop($box->items);

// The generated Frame/Block shape exercises a constructor-built stack and a
// method pop. Mutating the returned frame is observable through the old stack
// element, demonstrating object identity.
$frame = new PopFrame('root');
$frame->push_child('child');
$stack_child = $frame->stack[0];
$returned_frame = $frame->pop_child();
$returned_frame->name = 'child-mutated';
$returned_frame->block->name = 'child-block-mutated';

// PHP_INT_MAX can be popped and then automatically inserted again at the same
// numeric position.
// With MAX still present, one automatic reappend at MAX-1 remains valid; another
// append would overflow. The minimum boundary is a string key because the negative
// literal itself is outside this frontend's accepted literal subset.
$maximum = [9223372036854775807 => 'maximum'];
$maximum_value = array_pop($maximum);
$maximum[] = 'maximum-readded';
$maximum_pair = [9223372036854775807 => 'maximum', 9223372036854775806 => 'maximum-minus-one'];
$maximum_pair_value = array_pop($maximum_pair);
$maximum_pair[] = 'maximum-minus-one-readded';
$minimum = ['-9223372036854775808' => 'minimum'];
$minimum_value = array_pop($minimum);
$minimum[] = 'minimum-sentinel';

// Repeated assigned pops naturally drive a while condition, then an assigned
// pop selects both an if and a switch branch without embedding the builtin.
$loop = ['first', 'second', 'third'];
$loop_trace = '';
while (count($loop) !== 0) {
    $loop_value = array_pop($loop);
    $loop_trace = $loop_trace . $loop_value . ':';
}
$control = ['if-value', 'switch-value'];
$if_value = array_pop($control);
$if_result = 'wrong';
if ($if_value === 'switch-value') {
    $if_result = 'if-ok';
}
$switch_value = array_pop($control);
$switch_result = 'wrong';
switch ($switch_value) {
case 'if-value':
    $switch_result = 'switch-ok';
    break;
default:
    $switch_result = 'switch-wrong';
    break;
}

// Assigning the result over its target is valid only as the final observation:
// afterwards this local intentionally holds the popped scalar, not an array.
$alias_last = ['alias-value'];
$alias_last = array_pop($alias_last);

echo 'ordered:' . $ordered_value . ':' . count($ordered) . "\n";
$ignored = entries('ordered-entries', $ordered);
echo 'numeric:' . $negative_value . ':' . $sparse_value . ':' . count($discard) . "\n";
$ignored = entries('negative-entries', $negative);
$ignored = entries('sparse-entries', $sparse);
echo 'keys:' . $key_value . ':' . count($keys) . "\n";
$ignored = entries('keys-entries', $keys);
echo 'scalar-types:' . gettype($null_value) . ':' . gettype($blank_value) . ':' . gettype($false_value) . ':' . gettype($zero_value) . ':' . gettype($empty_value) . "\n";
echo 'copies:' . $source_copy[0]['nested']['value'] . ':' . $nested_value['nested']['value'] . ':' . $child->name . "\n";
echo 'parameter:' . $parameter_value . ':' . count($argument_values) . ':' . $argument_values[1] . "\n";
echo 'property:' . $property_value . ':' . count($box->items) . "\n";
echo 'frame:' . $stack_child->name . ':' . $stack_child->block->name . ':' . count($frame->stack) . "\n";
echo 'maximum:' . $maximum_value . ':' . $maximum_pair_value . ':' . $minimum_value . "\n";
$ignored = entries('maximum-entries', $maximum);
$ignored = entries('maximum-pair-entries', $maximum_pair);
$ignored = entries('minimum-entries', $minimum);
echo 'control:' . $loop_trace . ':' . $if_result . ':' . $switch_result . "\n";
echo 'alias:' . gettype($alias_last) . ':' . $alias_last . "\n";
