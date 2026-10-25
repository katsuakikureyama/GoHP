<?php
// Optional defaults are restricted to trailing untyped global parameters. This
// fixture distinguishes omitted defaults, explicit null, explicit values, and
// the existing clone-at-parameter-boundary behavior for an array value.
function label(array $items, $prefix = NULL, $suffix = null) {
    if (is_null($prefix)) {
        $prefix = 'default:';
    }
    if (is_null($suffix)) {
        $suffix = ':end';
    }
    return $prefix . count($items) . $suffix;
}

function copied_items($items, $unused = null): array {
    $items['local'] = 'changed';
    return $items;
}

function map_label($value, $suffix = null) {
    if (is_null($suffix)) {
        return $value . ':mapped';
    }
    return $value . $suffix;
}

function forward_label(array $items, $suffix = null) {
    return label($items, null, $suffix);
}

$source = ['item' => 'original'];
$omitted = label($source);
$explicit_null = label($source, null);
$explicit_values = label($source, 'given:', ':value');
$forwarded = forward_label($source);
$copied = copied_items($source);
$copied_explicit = copied_items($source, null);
$mapped = array_map('map_label', ['a', 'b']);

echo 'labels=' . $omitted . ':' . $explicit_null . ':' . $explicit_values . ':' . $forwarded . "\n";
echo 'copy=' . $copied['item'] . ':' . $copied['local'] . ':' . $copied_explicit['local'] . ':' . $source['item'] . ':' . (isset($source['local']) ? 'changed' : 'clean') . "\n";
echo 'mapped=' . $mapped[0] . ':' . $mapped[1] . "\n";
