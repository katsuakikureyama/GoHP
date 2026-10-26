<?php
// Differential coverage for the bounded ValueSlot control-flow lane.  Every
// selected function is nullary and has only scalar or one-level scalar-array
// locals.  The strtok functions deliberately remain on the legacy lane so the
// external consumer can also check shared Execution state.

function owned_nested_if_else() {
    $items = ['root' => 'base'];
    $outer = true;
    if ($outer) {
        $inner = false;
        if ($inner) {
            $items['branch'] = 'wrong';
        } else {
            $items['branch'] = 'inner-else';
        }
    } else {
        $items['branch'] = 'outer-else';
    }
    return $items['root'] . ':' . $items['branch'] . ':' . count($items);
}

function owned_else_only_update() {
    $items = ['root' => 'base'];
    if (false) {
        $items[] = 'then';
    } else {
        $items[] = 'else';
    }
    return $items['root'] . ':' . $items[0] . ':' . count($items);
}

function owned_while_zero() {
    $items = ['root' => 'kept'];
    $i = 0;
    while ($i !== 0) {
        $items[] = 'unreachable';
    }
    return $items['root'] . ':' . count($items);
}

function owned_while_steps() {
    $items = [];
    $i = 0;
    while ($i !== 3) {
        $items[] = $i;
        $i = $i + 1;
    }
    return $items[0] . ':' . $items[1] . ':' . $items[2] . ':' . count($items);
}

function owned_continue_break() {
    $items = [];
    $i = 0;
    while ($i !== 5) {
        $i = $i + 1;
        if ($i === 2) {
            continue;
        }
        $items[] = $i;
        if ($i === 3) {
            break;
        }
    }
    return $items[0] . ':' . $items[1] . ':' . count($items);
}

function owned_early_return() {
    $items = ['early' => 'yes'];
    $i = 0;
    while ($i !== 1) {
        if (true) {
            return $items;
        }
        $i = $i + 1;
    }
    $items['late'] = 'no';
    return $items;
}

function owned_copy_before_loop() {
    $source = ['keep' => 'source', 'seed' => 'copy'];
    $copy = $source;
    $i = 0;
    while ($i !== 2) {
        $source[] = $i;
        $i = $i + 1;
    }
    $copy['keep'] = 'copy';
    return $source['keep'] . ':' . $copy['keep'] . ':' . $source[0] . ':' . $copy['seed'] . ':' . count($source) . ':' . count($copy);
}

function owned_copy_inside_loop() {
    $items = ['base' => 'first'];
    $saved = [];
    $i = 0;
    while ($i !== 2) {
        $saved = $items;
        $items['base'] = 'step' . $i;
        $items[] = $i;
        $i = $i + 1;
    }
    return $saved['base'] . ':' . $saved[0] . ':' . count($saved) . ':' . $items['base'] . ':' . $items[1] . ':' . count($items);
}

function owned_body_only_locals() {
    $items = [];
    $i = 0;
    while ($i !== 2) {
        $word = $i === 0 ? 'zero' : 'one';
        $copy = ['value' => $word];
        $items[] = $copy['value'];
        $i = $i + 1;
    }
    return $items[0] . ':' . $items[1] . ':' . count($items);
}

function owned_returned_control_array() {
    $items = [-5 => 'negative', 'label' => null];
    $i = 0;
    while ($i !== 2) {
        if ($i === 0) {
            $items[] = 'zero';
        } else {
            $items['tail'] = 'end';
        }
        $i = $i + 1;
    }
    return $items;
}

function owned_nested_loops() {
    $items = [];
    $outer = 0;
    while ($outer !== 2) {
        $inner = 0;
        while ($inner !== 2) {
            $items[] = $outer . ':' . $inner;
            $inner = $inner + 1;
        }
        $outer = $outer + 1;
    }
    return $items[0] . ':' . $items[1] . ':' . $items[2] . ':' . $items[3] . ':' . count($items);
}

// The external consumer alone executes this path.  Its panic must not retain
// a local owner that changes a later ordinary call.
function owned_loop_overflow() {
    $items = [];
    $i = 0;
    while ($i !== 1) {
        $edge = 9223372036854775807;
        $items[] = $edge + 1;
        $i = $i + 1;
    }
    return $items;
}

// These stay legacy and make the complete generated source execution-aware.
function owned_control_state_start() {
    return strtok('first,second', ',');
}

function owned_control_state_next() {
    return strtok(',');
}

function legacy_nested_identity($value) {
    return $value;
}

echo 'nested=' . owned_nested_if_else() . "\n";
echo 'else=' . owned_else_only_update() . "\n";
echo 'zero=' . owned_while_zero() . "\n";
echo 'steps=' . owned_while_steps() . "\n";
echo 'flow=' . owned_continue_break() . "\n";
echo 'before=' . owned_copy_before_loop() . "\n";
echo 'inside=' . owned_copy_inside_loop() . "\n";
echo 'body=' . owned_body_only_locals() . "\n";
echo 'nested-loop=' . owned_nested_loops() . "\n";
$first = owned_returned_control_array();
$second = owned_returned_control_array();
$first['label'] = 'first';
$first[] = 'main';
echo 'returned=' . $first[-5] . ':' . $first[-4] . ':' . $first['label'] . ':' . $first[-3] . ':' . (isset($second['label']) ? 'set' : 'null') . "\n";
foreach ($second as $key => $value) {
    echo gettype($key) . ':' . $key . '=' . gettype($value) . ':' . $value . ';';
}
echo "\n";
$early = owned_early_return();
echo 'early=' . $early['early'] . ':' . count($early) . "\n";
echo 'state=' . owned_control_state_start() . ':' . owned_control_state_next() . "\n";
