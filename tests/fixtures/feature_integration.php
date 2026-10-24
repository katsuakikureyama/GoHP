<?php
// Original cross-stack fixture: legacy arrays feed newer array builtins,
// guarded string helpers and one-arm control flow on the same compiler.
function summarize_values($input) {
    $values = array_values($input);
    $keys = array_keys($input);
    $summary = '';
    foreach ($values as $index => $value) {
        if (is_string($value) && strlen(rtrim($value, '/')) === 4) {
            $summary = $summary . $keys[$index] . ':' . strtolower(rtrim($value, '/')) . ';';
        }
    }
    return array('summary' => $summary, 'size' => count($values));
}

$input = ArRaY(9 => 'WORD///', 'flag' => false, 'nested' => array('keep' => 'VALUE'), 'last' => 'TEST/');
$result = summarize_values($input);
echo $result['summary'] . $result['size'] . "\n";
$values = array_values($input);
echo array_keys($values[2])[0] . ':' . $values[2]['keep'] . "\n";
echo 'bool=' . is_bool($values[1]) . ':' . ((bool) array_values(array())) . "\n";
echo 'precedence=' . (1 . 2 + 3 === '15' && is_bool(true)) . "\n";
echo 'original=' . array_keys($input)[0] . ':' . $input['nested']['keep'] . "\n";
