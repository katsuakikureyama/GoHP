<?php
// Owned differential fixture for (array) through inferred-unknown values.
// It stays in the generated-public-property lane: no PHP general object casts,
// references, dynamic calls, or exception behavior are implied by this file.

class SnapshotChild {
    public $name = 'child-original';
}

class SnapshotBlock {
    // Declaration order is part of the generated object snapshot contract.
    public $title = 'block-original';
    public $nullable = null;
    public $zero = 0;
    public $blank = '';
    public $off = false;
    public $payload = null;
    public $child = null;
}

class SnapshotFrame {
    public $block = null;
}

class SnapshotEnvelope {
    public $frame = null;
}

class SnapshotCarrier {
    public $block = null;

    public function expose() {
        return $this->block;
    }
}

class EmptySnapshot {}

function snapshot_parameter($value) {
    return (array) $value;
}

function return_unknown($value) {
    return $value;
}

// array_pop is outside this fixture's bounded builtin domain. This produces
// the same inferred-unknown extracted-local shape as the Parser stack value.
function extract_top($items) {
    return $items[0];
}

function marker_once($value) {
    echo '<cast-once>';
    return $value;
}

$child = new SnapshotChild();
$block = new SnapshotBlock();
$block->child = $child;
$source = ['nested' => ['value' => 'source-original']];
$block->payload = $source;
$source['nested'] = ['value' => 'source-mutated'];

$frame = new SnapshotFrame();
$frame->block = $block;
$carrier = new SnapshotCarrier();
$carrier->block = $block;
$envelope = new SnapshotEnvelope();
$envelope->frame = $frame;

// Unannotated parameter, function/method return, extracted local, indexed
// local, and nested public properties all intentionally erase class precision.
$from_parameter = snapshot_parameter($block);
$from_function = (array) return_unknown($block);
$from_method = (array) $carrier->expose();
$blocks = [$block];
$from_index = (array) $blocks[0];
$stack = [$frame];
$stack_top = extract_top($stack);
$from_stack_property = (array) $stack_top->block;
$from_nested_property = (array) $envelope->frame->block;
$from_marker = (array) marker_once($block);
$empty = new EmptySnapshot();
$empty_snapshot = (array) return_unknown($empty);

// Each cast owns nested arrays, while child objects retain their identity.
$first = snapshot_parameter($block);
$first_payload = $first['payload'];
$first_payload['nested'] = ['value' => 'first-mutated'];
$first['payload'] = $first_payload;
$second = (array) return_unknown($block);
$second_payload = $second['payload'];
$second_payload['nested'] = ['value' => 'second-mutated'];
$second['payload'] = $second_payload;
$third = (array) $stack_top->block;
$child->name = 'child-mutated';
$after_child_mutation = (array) $block;

$keys = array_keys($from_stack_property);
$falsy = '';
if (is_null($from_stack_property['nullable'])) {
    $falsy = $falsy . 'N';
}
if ($from_stack_property['zero'] === 0) {
    $falsy = $falsy . 'Z';
}
if ($from_stack_property['blank'] === '') {
    $falsy = $falsy . 'B';
}
if ($from_stack_property['off'] === false) {
    $falsy = $falsy . 'F';
}

echo 'paths:' . $from_parameter['title'] . ':' . $from_function['title'] . ':' . $from_method['title'] . ':' . $from_index['title'] . ':' . $from_nested_property['title'] . ':' . $from_marker['title'] . "\n";
echo 'order:' . $keys[0] . ':' . $keys[1] . ':' . $keys[2] . ':' . $keys[3] . ':' . $keys[4] . ':' . $keys[5] . ':' . $keys[6] . "\n";
echo 'falsy:' . $falsy . ':empty=' . count($empty_snapshot) . "\n";
echo 'copies:' . $source['nested']['value'] . ':' . $first['payload']['nested']['value'] . ':' . $second['payload']['nested']['value'] . ':' . $third['payload']['nested']['value'] . "\n";
echo 'identity:' . $first['child']->name . ':' . $after_child_mutation['child']->name . "\n";
