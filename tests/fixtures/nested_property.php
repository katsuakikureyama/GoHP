<?php
// Owned differential fixture for nested property writes. It deliberately uses
// small Block/Frame-shaped classes rather than copying the WordPress source.
// The admitted lane is a bare local plus two or more declared property names.

function nested_once($value) {
    echo '<nested-rhs>';
    return $value;
}

function mutate_old_child($frame) {
    $frame->block->text = 'rhs-inner';
    $frame->block->items[] = 'rhs-item';
    return 'rhs-outer';
}

function nested_concat_mutate($frame) {
    $frame->block->text = 'rhs-left';
    return '+rhs-right';
}

function nested_append_mutate($frame) {
    $frame->block->items[] = 'rhs-inner-append';
    return 'rhs-outer-append';
}

class NestedBlock {
    public $text = null;
    public $items = null;
    public $innerContent = null;
    public $payload = null;
    public $child = null;
}

class NestedFrame {
    public $block = null;

    public function replace_block($replacement) {
        $this->block = $replacement;
        return 'replacement-rhs';
    }
}

$block = new NestedBlock();
$child = new NestedBlock();
$frame = new NestedFrame();
$frame->block = $block;
$block->child = $child;
$frame->block->items = [];
$frame->block->innerContent = [];

// Assignment supports ordinary PHP scalar values, then .= reads the final
// property only after the RHS is evaluated.
$frame->block->text = null;
$was_null = is_null($frame->block->text);
$frame->block->text = true;
$was_true = $frame->block->text;
$frame->block->text = 7;
$was_seven = $frame->block->text === 7;
$frame->block->text = 'seed';
$frame->block->text .= '-concat';

// Two distinct Frame instances share the same Block pointer.
$alias = new NestedFrame();
$alias->block = $frame->block;
$alias->block->text = 'shared';

// A three-property path keeps the child object's identity.
$frame->block->child->text = 'three-deep';

// list() intentionally leaves this local unknown to the semantic checker.
list($unknown_frame) = [$frame];
$unknown_frame->block->text = nested_once('from-list');
$frames = [$frame];
$from_array = $frames[0];
$from_array->block->text = 'from-array';

// Assigning arrays to a property and reading them back uses value semantics.
$payload = ['nested' => ['value' => 'stored']];
$frame->block->payload = $payload;
$payload['nested'] = ['value' => 'source-mutated'];
$payload_copy = $frame->block->payload;
$payload_copy['nested'] = ['value' => 'copy-mutated'];

// Appends retain null and clone nested array values.
$inner_value = ['value' => 'inner-stored'];
$frame->block->innerContent[] = null;
$frame->block->innerContent[] = $inner_value;
$inner_value['value'] = 'inner-mutated';

// Control-flow writes use the same target protocol.
$iteration = 0;
while ($iteration < 2) {
    if ($iteration === 0) {
        $frame->block->items[] = 'if-first';
    } else {
        $frame->block->items[] = 'if-second';
    }
    switch ('append') {
        case 'append':
            $frame->block->items[] = 'switch';
            break;
        default:
            $frame->block->items[] = 'wrong';
            break;
    }
    $iteration = $iteration + 1;
}

// RHS mutates the same object before the outer write. The outer assignment
// wins for text, while the RHS append remains observable.
$frame->block->text = mutate_old_child($frame);
$frame->block->text .= nested_concat_mutate($frame);
$concat_order = $frame->block->text;
$frame->block->items[] = nested_append_mutate($frame);
$before_replace = $frame->block;

// Writable chains are fetched after RHS, which replaces Frame::block.
$replacement = new NestedBlock();
$replacement->text = 'replacement-original';
$frame->block->text = $frame->replace_block($replacement);

$concat_old = new NestedBlock();
$concat_old->text = 'concat-old';
$concat_new = new NestedBlock();
$concat_new->text = 'concat-new';
$concat_frame = new NestedFrame();
$concat_frame->block = $concat_old;
$concat_frame->block->text .= $concat_frame->replace_block($concat_new);

$append_old = new NestedBlock();
$append_old->items = [];
$append_new = new NestedBlock();
$append_new->items = [];
$append_frame = new NestedFrame();
$append_frame->block = $append_old;
$append_frame->block->items[] = $append_frame->replace_block($append_new);

$flags = '';
if ($was_null) {
    $flags = $flags . 'N';
}
if ($was_true) {
    $flags = $flags . 'T';
}
if ($was_seven) {
    $flags = $flags . 'I';
}

echo 'scalars:' . $flags . "\n";
echo 'shared:' . $alias->block->text . ':' . $unknown_frame->block->text . "\n";
echo 'deep:' . $child->text . "\n";
echo 'arrays:' . $before_replace->payload['nested']['value'] . ':' . $payload['nested']['value'] . ':' . $payload_copy['nested']['value'] . "\n";
echo 'inner:' . (is_null($before_replace->innerContent[0]) ? 'null' : 'wrong') . ':' . $before_replace->innerContent[1]['value'] . "\n";
echo 'items:' . $before_replace->items[0] . ':' . $before_replace->items[1] . ':' . $before_replace->items[2] . ':' . $before_replace->items[3] . ':' . $before_replace->items[4] . "\n";
echo 'order:' . $before_replace->text . ':' . $frame->block->text . "\n";
echo 'updates:' . $concat_order . ':' . $before_replace->items[5] . ':' . $before_replace->items[6] . ':' . $from_array->block->text . "\n";
echo 'replace-updates:' . $concat_old->text . ':' . $concat_new->text . ':' . count($append_old->items) . ':' . $append_new->items[0] . "\n";
