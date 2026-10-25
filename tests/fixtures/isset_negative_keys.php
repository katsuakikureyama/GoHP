<?php
// Owned differential fixture for bounded negative-integer keys in isset().
function bool_word($value) {
    if ($value) {
        return 'T';
    }
    return 'F';
}
$values = [
    -5 => 'present',
    - /* a comment may separate the sign and magnitude */ 6 => 'comment',
    -9223372036854775807 => 'edge',
    0 => 'zero',
    -0 => 'negative-zero-replaces-zero',
    -7 => null,
    -8 => false,
    -9 => '',
    -10 => [],
    -11 => 0,
];
echo 'direct:' . bool_word(isset($values[-5])) . ':' . bool_word(isset($values[- /* comment */ 6])) . ':' . bool_word(isset($values[-9223372036854775807])) . ':' . bool_word(isset($values[-77])) . ':' . bool_word(isset($values[-7])) . ':' . bool_word(isset($values[-8])) . ':' . bool_word(isset($values[-9])) . ':' . bool_word(isset($values[-10])) . ':' . bool_word(isset($values[-11])) . "\n";
echo 'zero:' . bool_word(isset($values[-0])) . ':' . bool_word(isset($values[0])) . ':' . $values[0] . "\n";
$canonical = ['-5' => 'canonical-negative'];
$leading = ['-05' => 'leading-zero'];
echo 'normalization:' . bool_word(isset($canonical[-5])) . ':' . bool_word(isset($canonical['-5'])) . ':' . bool_word(isset($leading[-5])) . ':' . bool_word(isset($leading['-05'])) . "\n";
$first = '';
if (isset($values[-5])) {
    $first = 'present';
} elseif (isset($values[-9223372036854775807])) {
    $first = 'edge-wrong';
} else {
    $first = 'fallback-wrong';
}
$edge_only = [-9223372036854775807 => 'edge'];
$second = '';
if (isset($edge_only[-5])) {
    $second = 'present-wrong';
} else if (isset($edge_only[-9223372036854775807])) {
    $second = 'edge';
} else {
    $second = 'fallback-wrong';
}
$missing = [];
$third = '';
if (isset($missing[-5])) {
    $third = 'present-wrong';
} elseif (isset($missing[-9223372036854775807])) {
    $third = 'edge-wrong';
} else {
    $third = 'fallback';
}
echo 'branches:' . $first . ':' . $second . ':' . $third . "\n";
if (!isset($values[-77]) || strlen($values[-77]) === 0) {
    echo "short-circuit:missing-read-skipped\n";
}
$probe = [-5 => 'start'];
$before = count($probe);
$seen = isset($probe[-5]);
$keys = array_keys($probe);
$probe[] = 'source-tail';
$copy = $probe;
$copy[] = 'copy-tail';
$probe[] = 'source-next';
echo 'probe:' . bool_word($seen) . ':' . $before . ':' . count($keys) . ':' . $keys[0] . ':' . count($probe) . ':' . count($copy) . ':' . $probe[-4] . ':' . $probe[-3] . ':' . $copy[-4] . ':' . $copy[-3] . "\n";
