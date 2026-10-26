<?php
// Owned differential fixture: constructors are public, fixed-arity, untyped,
// and populate declared public properties. Direct __construct calls are
// deliberately absent; all construction flows through new.
function argument_marker($name, $value) {
    echo '<arg:' . $name . '>';
    return $value;
}

class ConstructorProbe {
    public $first;
    public $second;
    public $untouched;

    public function __CoNsTrUcT($first, $second) {
        echo '<constructor-body>';
        $this->first = $first;
        $this->second = $second;
    }
}

class BlockChild {
    public $name;

    public function __construct($name) {
        $this->name = $name;
    }
}

class BlockData {
    public $name;
    public $supports;
    public $child;
    public $untouched;

    public function __construct($name, $supports, $child) {
        $this->name = $name;
        $this->supports = $supports;
        $this->child = $child;
    }

    public function rename($name) {
        $this->name = $name;
        return $this->name;
    }
}

class PlainBlock {
    public $untouched;
}

$probe = new ConstructorProbe(argument_marker('first', 'A'), argument_marker('second', 'B'));
$source = ['nested' => ['child' => 'source']];
$child = new BlockChild('child-original');
$block = new BlockData('block-original', $source, $child);
$second = new BlockData('block-second', ['nested' => ['child' => 'second']], new BlockChild('child-second'));
$plain = new PlainBlock();

$source['nested'] = ['child' => 'source-mutated'];
$stored_before_copy = $block->supports;
$getter_copy = $block->supports;
$getter_copy['nested'] = ['child' => 'getter-copy'];
$stored_after_copy = $block->supports;

$snapshot = (array) $block;
$snapshot_keys = array_keys($snapshot);
$snapshot['name'] = 'cast-copy';
$snapshot['supports'] = ['nested' => ['child' => 'cast-copy']];
$block->name = 'object-mutated';
$child->name = 'child-mutated';
$after_object_mutation = (array) $block;
$method_result = $block->rename('method-name');
$after_method = (array) $block;

echo 'constructor:' . $probe->first . ':' . $probe->second . ':' . (is_null($probe->untouched) ? 'null' : 'wrong') . "\n";
echo 'fresh:' . $block->name . ':' . $second->name . ':' . $second->supports['nested']['child'] . ':' . (is_null($plain->untouched) ? 'null' : 'wrong') . "\n";
echo 'copies:' . $stored_before_copy['nested']['child'] . ':' . $source['nested']['child'] . ':' . $getter_copy['nested']['child'] . ':' . $stored_after_copy['nested']['child'] . "\n";
echo 'cast-order:' . $snapshot_keys[0] . ':' . $snapshot_keys[1] . ':' . $snapshot_keys[2] . ':' . $snapshot_keys[3] . "\n";
echo 'cast-copy:' . $snapshot['name'] . ':' . $snapshot['supports']['nested']['child'] . ':' . (is_null($snapshot['untouched']) ? 'null' : 'wrong') . "\n";
echo 'snapshot:' . $snapshot['name'] . ':' . $after_object_mutation['name'] . ':' . $snapshot['child']->name . ':' . $after_object_mutation['child']->name . "\n";
echo 'method:' . $method_result . ':' . $after_method['name'] . "\n";
