<?php
// Owned differential fixture: bare continue skips the remainder of its nearest
// foreach iteration while preserving effects that occurred before it.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function collect_kept(array $items, callable $skip): array {
    $kept = [];
    foreach ($items as $key => $value) {
        if ($skip($value, $key)) {
            continue;
        }
        $kept[] = $key . ':' . $value;
    }
    return $kept;
}

function has_kept(array $items, callable $skip): bool {
    foreach ($items as $key => $value) {
        if ($skip($value, $key)) {
            continue;
        }
        return true;
    }
    return false;
}

// Null, empty string, false, zero, and the string zero each take a distinct
// skipping branch. Only the truthy value reaches the post-branch side effect.
$values = ['null' => null, 'empty' => '', 'false' => false, 'zero' => 0, 'zero-text' => '0', 'yes' => 'yes'];
$skipped = '';
$kept = '';
$after_skip = '';
foreach ($values as $key => $value) {
    if ($value === null) {
        $skipped = $skipped . 'N';
        continue;
    } elseif ($value === '') {
        $skipped = $skipped . 'E';
        continue;
    } else if ($value === false || $value === 0 || $value === '0') {
        $skipped = $skipped . $key . ',';
        continue;
    } else {
        $kept = $kept . $key . ':' . $value . ',';
    }
    $after_skip = $after_skip . $key . ',';
}

// The three syntactic branches all use bare continue. Their effects remain,
// and no following statement in this foreach body needs to run.
$branch_trace = '';
foreach (['if' => 'one', 'elseif' => 'two', 'else' => 'three'] as $key => $value) {
    if ($key === 'if') {
        $branch_trace = $branch_trace . 'I';
        continue;
    } elseif ($key === 'elseif') {
        $branch_trace = $branch_trace . 'H';
        continue;
    } else {
        $branch_trace = $branch_trace . 'L';
        continue;
    }
}

// An inner continue advances only the inner foreach. The later outer continue
// runs after its inner loop has ended and still permits the next outer item.
$nested_trace = '';
$outer_tail = '';
$groups = [
    'first' => ['keep', 'skip', 'last'],
    'skip-outer' => ['inner'],
    'empty' => [],
    'last' => ['finish'],
];
foreach ($groups as $group_key => $group) {
    $nested_trace = $nested_trace . '[' . $group_key;
    foreach ($group as $value) {
        if ($value === 'skip') {
            $nested_trace = $nested_trace . ':S';
            continue;
        }
        $nested_trace = $nested_trace . ':' . $value;
    }
    $nested_trace = $nested_trace . ']';
    if ($group_key === 'skip-outer') {
        $outer_tail = $outer_tail . 'X';
        continue;
    }
    $outer_tail = $outer_tail . $group_key . ',';
}

// Writes made before continue survive. The later iteration performs both an
// append and a keyed write after its branch, proving the two paths are distinct.
$changes = ['start' => 'S'];
foreach (['skip' => false, 'keep' => true] as $key => $value) {
    if ($value === false) {
        $changes[] = 'before-skip';
        $changes['written'] = 'before';
        continue;
    }
    $changes[] = 'after-keep';
    $changes['written'] = 'after';
}

// foreach iterates its entry snapshot: the append and keyed write on the
// first pass change the array, but neither creates an extra pass nor changes
// the value already scheduled for the second pass.
$snapshot = ['first' => 'one', 'second' => 'two'];
$snapshot_seen = '';
$snapshot_after = '';
foreach ($snapshot as $key => $value) {
    $snapshot_seen = $snapshot_seen . $key . ':' . $value . ',';
    if ($key === 'first') {
        $snapshot[] = 'late';
        $snapshot['second'] = 'changed';
        continue;
    }
    $snapshot_after = $snapshot_after . $key . ',';
    $snapshot['after'] = $key;
}

// Typed array/bool returns compose a foreach continue with a capture-free
// callback. Empty input exercises both the loop-free array and bool paths.
$callback_items = ['left' => 'skip', 'middle' => 'keep', 'right' => 'stay'];
$callback_kept = collect_kept($callback_items, fn($value, $key) => $value === 'skip');
$callback_empty = collect_kept([], fn($value, $key) => $value === 'skip');
$callback_has = has_kept($callback_items, fn($value, $key) => $value === 'skip');
$callback_none = has_kept(['only' => 'skip'], fn($value, $key) => $value === 'skip');
$callback_empty_has = has_kept([], fn($value, $key) => $value === 'skip');

// A method owns its loop context too. Continue does not become a return.
class SkipRows {
    public function run($items) {
        $trace = '';
        foreach ($items as $value) {
            if ($value === 'skip') {
                continue;
            }
            $trace = $trace . $value . ',';
        }
        return $trace;
    }
}
$runner = new SkipRows();
$method_trace = $runner->run(['skip', 'kept', 'last']);

// Valid but unreachable source is checked without being emitted. Its array
// assignment must not erase the reachable false value in type analysis.
$state = [];
foreach ([1, 2] as $value) {
    $state = false;
    CoNtInUe;
    $state = [];
    echo 'UNREACHABLE';
}
$continue = 'ordinary-variable';

echo 'values=' . $skipped . ':' . $kept . ':' . $after_skip . "\n";
echo 'branches=' . $branch_trace . "\n";
echo 'nested=' . $nested_trace . ':' . $outer_tail . "\n";
echo 'changes=' . count($changes) . ':' . $changes[0] . ':' . $changes[1] . ':' . $changes['written'] . "\n";
echo 'snapshot=' . $snapshot_seen . ':' . $snapshot_after . ':' . count($snapshot) . ':' . $snapshot['second'] . ':' . $snapshot['after'] . "\n";
echo 'callback=' . count($callback_kept) . ':' . $callback_kept[0] . ':' . $callback_kept[1] . ':' . count($callback_empty) . ':' . bool_word($callback_has) . ':' . bool_word($callback_none) . ':' . bool_word($callback_empty_has) . "\n";
echo 'method=' . $method_trace . ':state=' . bool_word($state) . ':' . $continue . "\n";
