<?php
// Owned differential fixture: bounded named global array parameters only.
function forward_array(array $items): array {
    return copy_array($items);
}

function copy_array(array $items): array {
    return $items;
}

function ordered_arguments($prefix, ArRaY $items, $suffix): array {
    return ['prefix' => $prefix, 'items' => $items, 'suffix' => $suffix];
}

function array_flag(array $items): bool {
    return is_array($items);
}

function mutate_parameter(array $items): array {
    $items['changed'] = 'yes';
    $items[] = 'appended';
    return $items;
}

function mutate_child(array $child): array {
    $child['child'] = 'changed';
    return $child;
}

function marker_array($name, array $items): array {
    echo '<' . $name . '>';
    return $items;
}

function first_array(array $left, array $right): array {
    return $left;
}

function map_array_value(array $value): array {
    echo '<map>';
    return $value;
}

function scalar_after_entry(array $items): array {
    $items = 'reassigned';
    return ['value' => $items];
}

$empty = copy_array([]);
$source = [-5 => ['child' => 'original'], 'label' => true, 2 => 'tail'];
$forward = forward_array($source);
$ordered = ordered_arguments('left', $source, 'right');
$flag = array_flag($source);
$mutated = mutate_parameter($source);
$child = mutate_child($source[-5]);
$marked = first_array(marker_array('left', ['left' => 1]), marker_array('right', ['right' => 2]));
$mapped = array_map('map_array_value', [[-1 => 'one'], ['name' => 'two']]);
$mapped_empty = array_map('map_array_value', []);
$scalar = scalar_after_entry($source);

echo 'empty:' . count($empty) . "\n";
echo 'forward:' . $forward[-5]['child'] . ':' . ($forward['label'] ? 'true' : 'false') . ':' . $forward[2] . "\n";
echo 'ordered:' . $ordered['prefix'] . ':' . $ordered['items'][-5]['child'] . ':' . $ordered['suffix'] . "\n";
echo 'flag:' . ($flag === true ? 'true' : 'wrong') . "\n";
echo 'mutated:' . $mutated['changed'] . ':' . $mutated[3] . ':' . count($mutated) . "\n";
echo 'source:' . $source[-5]['child'] . ':' . (isset($source['changed']) ? 'changed' : 'clean') . ':' . count($source) . "\n";
echo 'child:' . $child['child'] . ':' . $source[-5]['child'] . "\n";
echo 'marked:' . $marked['left'] . "\n";
echo 'mapped:' . $mapped[0][-1] . ':' . $mapped[1]['name'] . ':' . count($mapped_empty) . "\n";
echo 'scalar:' . $scalar['value'] . "\n";
