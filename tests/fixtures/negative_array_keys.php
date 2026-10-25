<?php
function dump_entries($label, $items) {
    foreach ($items as $key => $value) {
        echo $label . ':' . (is_string($key) ? 'string' : 'int') . ':' . $key . '=' . $value . ',';
    }
    echo "\n";
    return '';
}

function identity_value($value) {
    return $value;
}

$short = ['-5' => 'short-negative', 'short-implicit'];
$legacy = array('-5' => 'legacy-negative', 'legacy-implicit');
$after_string = ['name' => 'prefix', '-5' => 'negative', 'implicit'];
$negative_order = ['-5' => 'first', '-7' => 'lower', '-3' => 'higher', '-3' => 'replacement', 'implicit'];
$cross_zero = ['-1' => 'minus-one', 'zero'];
$positive_then_negative = [3 => 'three', '-5' => 'negative', 'implicit'];
$canonical = ['-5' => 'integer', '-05' => 'string', 'implicit'];
$minimum = ['-9223372036854775808' => 'minimum', 'implicit'];

echo dump_entries('short', $short);
echo dump_entries('legacy', $legacy);
echo dump_entries('after-string', $after_string);
echo dump_entries('negative-order', $negative_order);
echo dump_entries('cross-zero', $cross_zero);
echo dump_entries('positive-then-negative', $positive_then_negative);
echo dump_entries('canonical', $canonical);
echo dump_entries('minimum', $minimum);

$source = ['-5' => 'source-negative', 'source-implicit'];
$copy = $source;
$mapped = array_map('identity_value', $copy);
$reindexed = array_values($mapped);
echo dump_entries('source', $source);
echo dump_entries('copy', $copy);
echo dump_entries('mapped', $mapped);
echo dump_entries('reindexed', $reindexed);
