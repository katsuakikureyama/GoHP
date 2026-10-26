<?php
// Owned differential fixture: braced while conditions are evaluated before
// every iteration. This covers only bare break/continue for while and
// foreach; it does not claim explicit levels, for/do loops, or switch.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function condition_marker($label, $current, $limit): bool {
    echo '<' . $label . ':' . $current . '>';
    return $current !== $limit;
}

function prefix_marker($value): bool {
    echo '<prefix:' . $value . '>';
    return str_starts_with($value, 'wp-elements-');
}

// This bounded lookup mirrors a small token selection helper only. It does
// not claim to run a complete WordPress or HTML-processor path.
function first_prefixed_while(array $tokens): array {
    $index = 0;
    $result = ['found' => false, 'token' => ''];
    while ($index !== count($tokens)) {
        $token = $tokens[$index];
        if (str_starts_with($token, 'wp-elements-')) {
            $result = ['found' => true, 'token' => $token];
            break;
        }
        $index = $index + 1;
    }
    return $result;
}

function has_while_stop(array $items): bool {
    $index = 0;
    $found = false;
    while ($index !== count($items)) {
        if ($items[$index] === 'stop') {
            $found = true;
            break;
        }
        $index = $index + 1;
    }
    return $found;
}

class WhilePicker {
    public function pick($tokens) {
        $picked = [];
        while (true) {
            $picked = first_prefixed_while($tokens);
            break;
        }
        return $picked;
    }
}

// Zero, one, and multiple iterations. The marker also proves the final false
// condition is observed after natural completion.
$zero = 0;
while (condition_marker('zero', $zero, 0)) {
    $zero = $zero + 1;
}
$one = 0;
while (condition_marker('one', $one, 1)) {
    $one = $one + 1;
}
$many = 0;
$many_trace = '';
WHILE (condition_marker('many', $many, 3)) {
    $many_trace = $many_trace . $many;
    $many = $many + 1;
}
echo "\ncounts=" . $zero . ':' . $one . ':' . $many . ':' . $many_trace . "\n";

// The right operand is reached only while the left condition is true. The
// first token fails the prefix condition, so the body never advances it.
$short_index = 0;
$short_tokens = ['plain'];
$short_body = '';
while (condition_marker('short-left', $short_index, count($short_tokens)) && prefix_marker($short_tokens[$short_index])) {
    $short_body = $short_body . 'body';
    $short_index = $short_index + 1;
}
echo "\nshort=" . $short_index . ':' . $short_body . "\n";
while (condition_marker('short-skip', 0, 0) && prefix_marker('UNREACHED')) {
    echo 'SHORT-DEAD';
}
echo "\n";

// Continue first updates the bounded counter, then the next condition must be
// re-evaluated. The dead tail must not be emitted.
$continue_index = 0;
$continue_trace = '';
while (condition_marker('continue', $continue_index, 3)) {
    if ($continue_index === 0) {
        $continue_trace = $continue_trace . 'C';
        $continue_index = $continue_index + 1;
        CoNtInUe;
        echo 'CONTINUE-DEAD';
    }
    $continue_trace = $continue_trace . $continue_index;
    $continue_index = $continue_index + 1;
}
echo "\ncontinue=" . $continue_trace . ':' . $continue_index . "\n";

// Break leaves the while immediately: unlike natural completion it must not
// perform a final condition evaluation.
$break_index = 0;
$break_trace = '';
while (condition_marker('break', $break_index, 3)) {
    $break_trace = $break_trace . $break_index;
    if ($break_index === 1) {
        $break_trace = $break_trace . 'B';
        break;
        echo 'BREAK-DEAD';
    }
    $break_index = $break_index + 1;
}
echo "\nbreak=" . $break_trace . ':' . $break_index . "\n";

// An inner while break leaves its enclosing foreach alive. Conversely, the
// inner foreach continue advances only that foreach, then the while tail runs.
$groups = [
    'one' => ['left', 'stop', 'later'],
    'two' => ['again'],
    'empty' => [],
];
$nested_trace = '';
foreach ($groups as $group_key => $group) {
    $group_index = 0;
    $nested_trace = $nested_trace . '[' . $group_key;
    while ($group_index !== count($group)) {
        $nested_trace = $nested_trace . ':' . $group[$group_index];
        if ($group[$group_index] === 'stop') {
            $nested_trace = $nested_trace . ':B';
            break;
        }
        $group_index = $group_index + 1;
    }
    $nested_trace = $nested_trace . ']';
}

$outer_index = 0;
$foreach_in_while = '';
$rows = [
    ['skip', 'a'],
    ['b'],
];
while ($outer_index !== count($rows)) {
    foreach ($rows[$outer_index] as $value) {
        if ($value === 'skip') {
            continue;
            echo 'FOREACH-CONTINUE-DEAD';
        }
        $foreach_in_while = $foreach_in_while . $value;
    }
    $foreach_in_while = $foreach_in_while . '|';
    $outer_index = $outer_index + 1;
}
echo 'nested=' . $nested_trace . ':' . $foreach_in_while . "\n";

// Array parameters remain caller-owned. The callback and method compose the
// typed while helpers without widening this fixture into a source-application
// claim.
$source_tokens = ['plain', 'wp-elements-card', 'wp-elements-later'];
$source_count = count($source_tokens);
$first = first_prefixed_while($source_tokens);
$picker = new WhilePicker();
$method_first = $picker->pick(['plain', 'wp-elements-method']);
$flags = array_map('has_while_stop', [['stop'], ['keep'], []]);
echo 'tokens=' . bool_word($first['found']) . ':' . $first['token']
    . ':' . $source_tokens[0] . ':' . count($source_tokens) . ':' . $source_count
    . ':' . bool_word($method_first['found']) . ':' . $method_first['token'] . "\n";
echo 'callback=' . bool_word($flags[0]) . ':' . bool_word($flags[1]) . ':' . bool_word($flags[2]) . "\n";
