<?php
// Owned differential fixture for the bounded string-switch lane. Selectors and
// case labels are nonnumeric strings; this does not claim PHP's coercive
// general switch comparison, numeric labels, explicit break/continue levels,
// or a complete application parser.

function switch_marker($value) {
    echo '<selector:' . $value . '>';
    return $value;
}

function switch_return_arm($value) {
    switch ($value) {
        case 'first-return':
            return 'first';
        case 'empty-return':
        case 'second-return':
            return 'second';
        default:
            return 'default';
    }
}

function switch_conditional_return($value) {
    switch ($value) {
        case 'conditional':
            if (true) { return 'yes'; } else { return 'no'; }
        default:
            return 'fallback';
    }
}

class StringSwitchMethod {
    public function pick($value) {
        switch ($value) {
            case 'method-first':
                return 'method-first';
            case 'method-empty':
            case 'method-second':
                return 'method-second';
            default:
                return 'method-default';
        }
    }
}

// A default may appear before a later matching case. It participates only
// after no case label matches, regardless of its source position.
$default_before_later = '';
switch ('later-match') {
    default:
        $default_before_later = 'wrong-default';
        break;
    case 'later-match':
        $default_before_later = 'later';
        break;
}

// A selected default falls through into the following case when it has no
// break. A switch with no default leaves this predeclared local intact.
$default_fallthrough = '';
switch ('missing-label') {
    default:
        $default_fallthrough = $default_fallthrough . 'D';
    case 'after-default':
        $default_fallthrough = $default_fallthrough . 'A';
        break;
}
$no_default_miss = 'kept';
switch ('not-present') {
    case 'present':
        $no_default_miss = 'wrong';
        break;
}

// The first duplicate case is selected. Its ordinary fallthrough reaches the
// later duplicate label as an ordinary statement position.
$duplicate_trace = '';
switch ('duplicate') {
    case 'duplicate':
        $duplicate_trace = $duplicate_trace . 'first,';
    case 'middle':
        $duplicate_trace = $duplicate_trace . 'middle,';
    case 'duplicate':
        $duplicate_trace = $duplicate_trace . 'later,';
        break;
    default:
        $duplicate_trace = 'wrong-default';
        break;
}

// An empty case label shares the following body. Matching is byte and
// case-sensitive for this lane.
$empty_label = '';
switch ('empty-label') {
    case 'empty-label':
    case 'joined-label':
        $empty_label = 'joined';
        break;
    default:
        $empty_label = 'wrong';
        break;
}
$case_sensitive = '';
switch ('case-byte-é') {
    case 'CASE-BYTE-é':
        $case_sensitive = 'wrong-case';
        break;
    case 'case-byte-É':
        $case_sensitive = 'wrong-byte';
        break;
    case 'case-byte-é':
        $case_sensitive = 'matched-bytes';
        break;
    default:
        $case_sensitive = 'wrong-default';
        break;
}

// The selector expression runs once, before the matching scan begins.
$once = '';
switch (switch_marker('once-only')) {
    case 'once-only':
        $once = 'hit';
        break;
    default:
        $once = 'wrong';
        break;
}

// A variable predeclared before the switch is written by each arm. This keeps
// the branch scope deliberately conservative.
$branch_state = 'before';
switch ('branch-second') {
    case 'branch-first':
        $branch_state = 'first';
        break;
    case 'branch-second':
        $branch_state = 'second';
        break;
    default:
        $branch_state = 'default';
        break;
}

// An inner while owns its continue and break. The following switch statement
// remains reachable after the loop break, then the switch's own break exits
// only the switch.
$inner_trace = '';
$inner_index = 0;
switch ('inner-loop') {
    case 'inner-loop':
        while ($inner_index < 4) {
            $inner_index = $inner_index + 1;
            if ($inner_index === 2) {
                $inner_trace = $inner_trace . 'C';
                continue;
            }
            $inner_trace = $inner_trace . $inner_index;
            if ($inner_index === 3) {
                break;
            }
        }
        $inner_trace = $inner_trace . ':switch-tail';
        break;
    default:
        $inner_trace = 'wrong';
        break;
}

// A break inside a case exits the switch. The outer while chooses separately
// to stop after the switch has returned control to it.
$outer_values = ['go', 'stop', 'late'];
$outer_index = 0;
$outer_trace = '';
while ($outer_index < count($outer_values)) {
    $outer_selector = $outer_values[$outer_index];
    switch ($outer_selector) {
        case 'go':
            $outer_trace = $outer_trace . 'G';
            break;
        case 'stop':
            $outer_trace = $outer_trace . 'S';
            break;
        default:
            $outer_trace = $outer_trace . 'X';
            break;
    }
    if ($outer_selector === 'stop') {
        break;
    }
    $outer_index = $outer_index + 1;
}

// Nested switches each own their nearest break.
$nested_trace = '';
switch ('outer') {
    case 'outer':
        switch ('inner') {
            case 'inner':
                $nested_trace = $nested_trace . 'I';
                break;
            default:
                $nested_trace = $nested_trace . 'wrong-inner';
                break;
        }
        $nested_trace = $nested_trace . 'O';
        break;
    default:
        $nested_trace = 'wrong-outer';
        break;
}

// Selector evaluation shares the existing request-local strtok state.
$token_trace = '';
switch (strtok('first-token,second-token', ',')) {
    case 'first-token':
        $token_trace = 'first';
        break;
    default:
        $token_trace = 'wrong';
        break;
}
$token_trace = $token_trace . ':' . strtok(',');

$method = new StringSwitchMethod();
echo 'dispatch=' . $default_before_later . ':' . $default_fallthrough . ':' . $no_default_miss . ':' . $duplicate_trace . "\n";
echo 'labels=' . $empty_label . ':' . $case_sensitive . ':' . $once . ':' . $branch_state . "\n";
echo 'loops=' . $inner_trace . ':' . $inner_index . ':' . $outer_trace . ':' . $outer_index . ':' . $nested_trace . "\n";
echo 'returns=' . switch_return_arm('first-return') . ':' . switch_return_arm('empty-return') . ':' . switch_return_arm('other-return') . ':' . $method->pick('method-empty') . ':' . $method->pick('other-method') . "\n";
echo 'execution=' . $token_trace . "\n";
echo 'conditional=' . switch_conditional_return('conditional') . "\n";
