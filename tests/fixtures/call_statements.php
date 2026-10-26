<?php
// Owned differential fixture: a supported call used as a statement discards
// only its result. Its argument evaluation, call side effects, runtime guards,
// and enclosing control-flow behavior are retained. This does not cover
// arbitrary expression statements, chained calls, or direct constructor calls.

function mark($value) {
    echo '<' . $value . '>';
    return $value;
}

function scalar_result() {
    return 'scalar';
}

function array_result(): array {
    return ['array'];
}

function bool_result(): bool {
    return true;
}

class ReturnedObject {
}

function object_result() {
    return new ReturnedObject();
}

function arrow_result() {
    return fn($value) => $value;
}

function constructor_mark($value) {
    echo '<constructor:' . $value . '>';
    return $value;
}

function named_map_mark($value) {
    echo '<named:' . $value . '>';
    return $value;
}

function arrow_map_mark($value) {
    $callback = fn($item) => named_map_mark('arrow-' . $item);
    $callback($value);
    return $value;
}

class CallRecorder {
    public $trace = '';

    public function __construct($label) {
        constructor_mark($label);
        $this->trace = 'constructed';
    }

    public function record($value) {
        $this->trace = $this->trace . ':' . $value;
        return $this->trace;
    }
}

// Function, builtin, arrow, and bare generated-object method calls all discard
// their return values. The first and last calls visibly retain side effects;
// the return-only calls deliberately have no output.
mark('function');
mark('outer-' . mark('argument'));
strlen('builtin-discarded');
scalar_result();
array_result();
object_result();
bool_result();
arrow_result();
$defined_arrow = fn($value) => $value;
$defined_arrow('arrow-discarded');

$recorder = new CallRecorder('ready');
$recorder->record('method');
echo "\nmethod=" . $recorder->trace . "\n";

// array_map itself is a discarded builtin result. Its literal named callbacks
// run for every input; an adapter invokes an arrow as a standalone statement.
array_map('named_map_mark', ['one', 'two']);
array_map('arrow_map_mark', ['three', 'four']);
echo "\n";

// A statement call belongs to the branch that executes. Continue and break
// suppress statement calls after them just like other statements.
if (true) {
    mark('if');
} else {
    mark('if-dead');
}

foreach (['skip', 'keep', 'stop', 'later'] as $value) {
    if ($value === 'skip') {
        mark('foreach-skip');
        continue;
        mark('foreach-continue-dead');
    }
    mark('foreach-' . $value);
    if ($value === 'stop') {
        mark('foreach-break');
        break;
        mark('foreach-break-dead');
    }
}

$index = 0;
while ($index !== 3) {
    $index = $index + 1;
    if ($index === 1) {
        mark('while-continue');
        continue;
        mark('while-continue-dead');
    }
    mark('while-' . $index);
    if ($index === 2) {
        mark('while-break');
        break;
        mark('while-break-dead');
    }
}
echo "\n";

// The initializing strtok result is discarded, but it still establishes the
// request-local scan consumed by the following continuation.
strtok('a,b', ',');
$next_token = strtok(',');
echo 'strtok=' . $next_token . "\n";
