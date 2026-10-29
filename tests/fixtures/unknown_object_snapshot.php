<?php
// Owned differential fixture for the intentionally narrow unknown-object
// snapshot path. The cast operand reaches this point through a stack index, an
// untyped frame property and an untyped global return, like the parser stack
// property path. array_pop is a separate unimplemented builtin; this fixture
// does not certify stack mutation or the original parser.

class SnapshotChild {
    public $name;

    public function __construct($name) {
        $this->name = $name;
    }
}

class SnapshotBlock {
    public $name;
    public $payload;
    public $child;
    public $nullable;

    public function __construct($name, $payload, $child) {
        $this->name = $name;
        $this->payload = $payload;
        $this->child = $child;
    }
}

class SnapshotFrame {
    public $block;

    public function __construct($block) {
        $this->block = $block;
    }
}

class SnapshotTrace {
    public $calls = 0;
}

// Both parameters and the global return are deliberately untyped. The body
// makes the receiver observable once before the cast's actual-domain guard.
function stack_block_once($stack, SnapshotTrace $trace) {
    $trace->calls = $trace->calls + 1;
    $frame = $stack[0];
    return $frame->block;
}

$source = ['outer' => ['value' => 'source']];
$child = new SnapshotChild('child-original');
$block = new SnapshotBlock('block-original', $source, $child);
$stack = [new SnapshotFrame($block)];
$trace = new SnapshotTrace();

// stack index -> frame->block produces an inferred-unknown object. The runtime
// snapshot must be fresh and declaration ordered, while nested arrays copy and
// child object identity remains shared.
$snapshot = (array) stack_block_once($stack, $trace);
$keys = array_keys($snapshot);
$snapshot['name'] = 'snapshot-name';
$snapshot['payload'] = ['outer' => ['value' => 'snapshot']];
$block->name = 'block-mutated';
$block->payload = ['outer' => ['value' => 'block-mutated']];
$child->name = 'child-mutated';
$after = (array) $block;

echo 'once:' . $trace->calls . ':' . count($stack) . "\n";
echo 'order:' . $keys[0] . ':' . $keys[1] . ':' . $keys[2] . ':' . $keys[3] . "\n";
echo 'copy:' . $snapshot['name'] . ':' . $snapshot['payload']['outer']['value'] . ':' . $after['name'] . ':' . $after['payload']['outer']['value'] . "\n";
echo 'identity:' . $snapshot['child']->name . ':' . $after['child']->name . "\n";
echo 'null:' . (is_null($snapshot['nullable']) ? 'null' : 'wrong') . ':' . (is_null($after['nullable']) ? 'null' : 'wrong') . "\n";
