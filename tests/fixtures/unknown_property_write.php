<?php
// Owned differential fixture for property writes through locals whose precise
// generated class is intentionally no longer known to the semantic checker.
// It uses only declared public slots on real generated objects.

class UnknownWriteBox {
    public $offset = null;
    public $items = null;
    public $nullable = null;
    public $zero = 0;
    public $blank = '';
    public $off = false;
    public $payload = null;
}

class ThisWriteBox {
    public $value = null;

    public function replace($value) {
        $this->value = $value;
        return $this->value;
    }
}

// The parameter has no class annotation. Its property writes must still be
// guarded by the runtime object's generated PropertyObject implementation.
function write_from_untyped_parameter($item) {
    $item->offset = 'parameter';
    $item->items[] = 'parameter-append';
    return $item;
}

// The outer append must fetch items after this RHS replaces the property.
function replace_items(UnknownWriteBox $item) {
    $item->items = ['rhs-replaced'];
    return 'outer-append';
}

$object = new UnknownWriteBox();
$object->items = ['initial'];
$list = [$object];
$item = $list[0];
$item->offset = 'array-extracted';

$alias = $item;
$alias->offset = 'alias-shared';
$other = new UnknownWriteBox();
$other->offset = 'other-only';

$item = write_from_untyped_parameter($item);
$parameter_items = $item->items;

// Each declared slot begins as null or a falsy value, then receives a normal
// scalar through the inferred-unknown local.
$item->nullable = 'from-null';
$item->zero = 'from-zero';
$item->blank = 'from-blank';
$item->off = 'from-false';

$source = ['nested' => ['value' => 'source']];
$item->payload = $source;
$source['nested'] = ['value' => 'source-mutated'];
$payload_copy = $item->payload;
$payload_copy['nested'] = ['value' => 'copy-mutated'];

$item->items[] = replace_items($item);

// Assignment through a branch and loop deliberately widens these locals.
$if_object = new UnknownWriteBox();
$if_list = [$if_object];
$if_item = $other;
if (true) {
    $if_item = $if_list[0];
} else {
    $if_item = $other;
}
$if_item->offset = 'if-write';

$while_object = new UnknownWriteBox();
$while_list = [$while_object];
$while_item = $other;
$loop = 0;
while ($loop < 1) {
    $while_item = $while_list[0];
    $loop = $loop + 1;
}
$while_item->offset = 'while-write';

// $this retains its established known-receiver path.
$method = new ThisWriteBox();
$method_result = $method->replace('method-write');

echo 'identity:' . $object->offset . ':' . $alias->offset . ':' . $other->offset . "\n";
echo 'parameter:' . $parameter_items[0] . ':' . $parameter_items[1] . "\n";
echo 'falsy:' . $item->nullable . ':' . $item->zero . ':' . $item->blank . ':' . $item->off . "\n";
echo 'copies:' . $item->payload['nested']['value'] . ':' . $source['nested']['value'] . ':' . $payload_copy['nested']['value'] . "\n";
echo 'append:' . $item->items[0] . ':' . $item->items[1] . "\n";
echo 'flow:' . $if_object->offset . ':' . $while_object->offset . ':' . $loop . "\n";
echo 'method:' . $method_result . ':' . $method->value . "\n";
