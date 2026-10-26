<?php
// Owned differential fixture: bare break exits only its nearest foreach and
// keeps effects that were made before the break. It does not claim support for
// explicit break levels or non-foreach loops.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function first_prefixed_token(array $tokens): array {
    $result = ['found' => false, 'token' => ''];
    foreach ($tokens as $token) {
        if (str_starts_with($token, 'wp-elements-')) {
            $result = ['found' => true, 'token' => $token];
            break;
        }
    }
    return $result;
}

function has_middle_break(array $items): bool {
    $found = false;
    foreach ($items as $value) {
        if ($value === 'stop') {
            $found = true;
            break;
        }
    }
    return $found;
}

function copied_after_break(array $items): array {
    foreach ($items as $key => $value) {
        $items['visited'] = $key . ':' . $value;
        break;
    }
    return $items;
}

// The break position is first, middle, absent, or in an empty input. Effects
// before the break remain; the statement after it must never run.
$first_trace = '';
foreach (['first', 'later'] as $value) {
    $first_trace = $first_trace . $value . ',';
    if ($value === 'first') {
        $first_trace = $first_trace . 'B,';
        BrEaK;
        echo 'FIRST-DEAD';
    }
    $first_trace = $first_trace . 'tail,';
}

$middle_trace = '';
foreach (['left', 'middle', 'right'] as $value) {
    $middle_trace = $middle_trace . $value . ',';
    if ($value === 'middle') {
        $middle_trace = $middle_trace . 'B,';
        break;
        echo 'MIDDLE-DEAD';
    }
    $middle_trace = $middle_trace . 'tail,';
}

$none_trace = '';
foreach (['left', 'right'] as $value) {
    $none_trace = $none_trace . $value . ',';
    if ($value === 'missing') {
        break;
    }
    $none_trace = $none_trace . 'tail,';
}

$empty_trace = '';
foreach ([] as $value) {
    $empty_trace = $empty_trace . $value;
    break;
}

// Braced if, elseif, and else each take their own bare break path.
$branch_if = '';
foreach (['if', 'later'] as $value) {
    if ($value === 'if') {
        $branch_if = $branch_if . 'I';
        break;
    } elseif ($value === 'elseif') {
        $branch_if = $branch_if . 'H';
        break;
    } else {
        $branch_if = $branch_if . 'L';
        break;
    }
}
$branch_elseif = '';
foreach (['elseif', 'later'] as $value) {
    if ($value === 'if') {
        $branch_elseif = $branch_elseif . 'I';
        break;
    } elseif ($value === 'elseif') {
        $branch_elseif = $branch_elseif . 'H';
        break;
    } else {
        $branch_elseif = $branch_elseif . 'L';
        break;
    }
}
$branch_else = '';
foreach (['else', 'later'] as $value) {
    if ($value === 'if') {
        $branch_else = $branch_else . 'I';
        break;
    } elseif ($value === 'elseif') {
        $branch_else = $branch_else . 'H';
        break;
    } else {
        $branch_else = $branch_else . 'L';
        break;
    }
}

// An inner break must leave the outer foreach running, while preserving the
// inner write that occurred before it.
$nested_trace = '';
$outer_tail = '';
$groups = [
    'one' => ['left', 'stop', 'later'],
    'two' => ['again'],
    'empty' => [],
];
foreach ($groups as $group_key => $group) {
    $nested_trace = $nested_trace . '[' . $group_key;
    foreach ($group as $value) {
        $nested_trace = $nested_trace . ':' . $value;
        if ($value === 'stop') {
            $nested_trace = $nested_trace . ':B';
            break;
            echo 'INNER-DEAD';
        }
    }
    $nested_trace = $nested_trace . ']';
    $outer_tail = $outer_tail . $group_key . ',';
}

// The loop receives its own array parameter. Mutating it before break must
// not alter the caller's array, and its pre-break mutation is retained in the
// returned copy.
$source = ['first' => 'one', 'second' => 'two'];
$copied = copied_after_break($source);

// A method has an independent foreach context. Its break is not a return.
class BreakRows {
    public function run($items) {
        $trace = '';
        foreach ($items as $value) {
            $trace = $trace . $value . ',';
            if ($value === 'stop') {
                break;
            }
        }
        return $trace . 'method-tail,';
    }
}
$runner = new BreakRows();
$method_trace = $runner->run(['keep', 'stop', 'later']);

// This bounded composition selects the first owned token with the WordPress
// element prefix; it is not a claim to execute the full WordPress source.
$tokens = ['plain', 'wp-elements-card', 'wp-elements-later'];
$first_token = first_prefixed_token($tokens);
$no_token = first_prefixed_token(['plain', 'component']);

$control = '';
foreach (['skip', 'keep', 'stop', 'later'] as $value) {
    if ($value === 'skip') { continue; }
    if ($value === 'stop') { break; }
    $control = $control . $value;
}
$snapshot = ['first' => 'one', 'second' => 'two'];
$snapshot_trace = '';
foreach ($snapshot as $key => $value) {
    $snapshot['second'] = 'changed';
    $snapshot_trace = $snapshot_trace . $value;
    if ($key === 'second') { break; }
}
$flags = array_map('has_middle_break', [['stop'], ['keep'], []]);
$break = 'variable-name';
echo 'control=' . $control . ':' . $snapshot_trace . ':' . $snapshot['second'] . ':' . $break . "\n";
echo 'callback=' . bool_word($flags[0]) . ':' . bool_word($flags[1]) . ':' . bool_word($flags[2]) . "\n";

echo 'positions=' . $first_trace . ':' . $middle_trace . ':' . $none_trace . ':' . $empty_trace . "\n";
echo 'branches=' . $branch_if . ':' . $branch_elseif . ':' . $branch_else . "\n";
echo 'nested=' . $nested_trace . ':' . $outer_tail . "\n";
echo 'copy=' . $source['second'] . ':' . count($source) . ':' . $copied['visited'] . ':' . count($copied) . "\n";
echo 'typed=' . bool_word(has_middle_break(['left', 'stop'])) . ':' . $method_trace . "\n";
echo 'tokens=' . bool_word($first_token['found']) . ':' . $first_token['token'] . ':' . bool_word($no_token['found']) . ':' . $no_token['token'] . "\n";
