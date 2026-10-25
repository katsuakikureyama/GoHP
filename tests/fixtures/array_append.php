<?php
function dump_entries($label, $items) {
    foreach ($items as $key => $value) {
        echo $label . ':' . (is_string($key) ? 'string' : 'int') . ':' . $key . '=' . $value . ',';
    }
    echo "\n";
    return '';
}

function append_in_function($items) {
    $items[] = 'function-tail';
    return $items;
}

function marker($value) {
    echo 'M';
    return $value;
}

function identity_value($value) {
    return $value;
}

$empty = [];
$empty[] = 'zero';
$empty[] = 'one';
$string_only = ['label' => 'value'];
$string_only[] = 'first';
$negative = ['-5' => 'negative'];
$negative[] = 'next';
echo dump_entries('empty', $empty);
echo dump_entries('string-only', $string_only);
echo dump_entries('negative', $negative);

$written = ['label' => 'prefix'];
$written['-5'] = 'negative';
$written[] = 'after-write';
$written['-9'] = 'lower';
$written['-2'] = 'higher';
$written['-2'] = 'replacement';
$written[] = 'minus-one';
$written[] = 'zero';
echo dump_entries('written', $written);
$minimum = ['-9223372036854775808' => 'minimum'];
$minimum[] = 'next';
echo dump_entries('minimum', $minimum);
$maximum = [9223372036854775806 => 'before-maximum'];
$maximum[] = 'maximum';
echo dump_entries('maximum', $maximum);

$copied = $negative;
$copied[] = 'copy-tail';
$mapped = array_map('identity_value', $negative);
$mapped[] = 'map-tail';
$keys = array_keys($negative);
$keys[] = 'keys-tail';
$values = array_values($negative);
$values[] = 'values-tail';
echo dump_entries('original', $negative);
echo dump_entries('copied', $copied);
echo dump_entries('mapped', $mapped);
echo dump_entries('keys', $keys);
echo dump_entries('values', $values);

$source = ['-5' => 'source-negative'];
$returned = append_in_function($source);
$returned[] = 'caller-tail';
echo dump_entries('source', $source);
echo dump_entries('returned', $returned);

$rhs = ['deep' => 'rhs'];
$receiver = ['base' => 'base'];
$receiver[] = $rhs;
$rhs['deep'] = 'changed';
echo $receiver[0]['deep'] . ':' . $rhs['deep'] . "\n";

$self = ['entry' => 'before'];
$self[] = $self;
$self['entry'] = 'after';
echo $self[0]['entry'] . ':' . $self['entry'] . "\n";

$marked = [];
$marked[] = marker('rhs');
echo $marked[0] . "\n";

$iterated = ['start' => 'first'];
$seen = '';
foreach ($iterated as $key => $value) {
    $seen = $seen . $key . ',';
    $iterated[] = 'late';
}
echo $seen . ':' . count($iterated) . ':' . $iterated[0] . "\n";
