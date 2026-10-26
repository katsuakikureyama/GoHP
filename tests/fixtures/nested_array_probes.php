<?php
// Owned differential fixture: nested isset() and ?? use only defined roots
// and integer/string keys. Reached intermediates are arrays or null.
function bool_word($value) {
    if ($value) {
        return 'T';
    }
    return 'F';
}

function fallback_marker($value) {
    echo '<' . $value . '>';
    return $value;
}

$values = [
    'branch' => [
        'false' => false,
        'zero' => 0,
        'empty' => '',
        'null' => null,
        'nested' => ['value' => 'deep'],
    ],
    'nil-middle' => null,
    'present' => 'present-value',
];
$empty_root = [];
$null_root = null;
$branch_key = 'branch';
$false_key = 'false';
$missing_key = 'missing';

$nested_false = isset($values[$branch_key][$false_key]);
$nested_null = isset($values['branch']['null']);
$nested_missing = isset($values['branch'][$missing_key]);
$nested_deep = isset($values['branch']['nested']['value']);
$empty_path = isset($empty_root['missing']['later']);
$nil_path = isset($values['nil-middle']['later']);
$null_root_path = isset($null_root['missing']['later']);
$multi_present = isset($values['branch']['false'], $values['branch']['zero'], $values['branch']['empty']);
$multi_short = isset($values['branch']['null'], $values['branch']['nested']['value']);
$multi_nil = isset($values['nil-middle']['later'], $values['branch']['false']);

$coalesce_false = $values['branch']['false'] ?? 'fallback';
$coalesce_zero = $values['branch']['zero'] ?? 'fallback';
$coalesce_empty = $values['branch']['empty'] ?? 'fallback';
$coalesce_null = $values['branch']['null'] ?? 'fallback';
$coalesce_missing = $values['branch']['missing'] ?? 'fallback';
$bare_false = false;
$bare_null = null;
$bare_false_default = $bare_false ?? 'fallback';
$bare_null_default = $bare_null ?? 'fallback';
$first = ['missing' => null];
$second = ['present' => 'chosen'];
$chain = $first['missing'] ?? $second['present'] ?? 'fallback';
$fallback_once = $values['branch']['missing'] ?? fallback_marker('fallback-once');
$fallback_skip = $values['present'] ?? fallback_marker('skip');
$or_precedence = $values['branch']['empty'] ?? false || true;
$ternary_precedence = $values['branch']['false'] ?? true ? 'wrong' : 'chosen';

echo 'isset:' . bool_word($nested_false) . ':' . bool_word($nested_null) . ':' . bool_word($nested_missing) . ':' . bool_word($nested_deep) . ':' . bool_word($empty_path) . ':' . bool_word($nil_path) . ':' . bool_word($null_root_path) . "\n";
echo 'multi:' . bool_word($multi_present) . ':' . bool_word($multi_short) . ':' . bool_word($multi_nil) . "\n";
echo 'coalesce:' . bool_word($coalesce_false) . ':' . $coalesce_zero . ':[' . $coalesce_empty . ']:' . $coalesce_null . ':' . $coalesce_missing . "\n";
echo 'bare:' . bool_word($bare_false_default) . ':' . $bare_null_default . "\n";
echo 'chain:' . $chain . "\n";
echo 'fallback:' . $fallback_once . ':' . $fallback_skip . "\n";
echo 'precedence:[' . $or_precedence . ']:' . $ternary_precedence . "\n";
