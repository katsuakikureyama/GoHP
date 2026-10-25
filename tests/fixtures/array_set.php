<?php
// A narrow array-element-write fixture. It compares the observed PHP order of
// the marker calls, but does not claim that order for arbitrary PHP lvalues.

function marker_key($key) {
    echo 'K' . $key;
    return $key;
}

function marker_value($value) {
    echo 'V' . $value;
    return $value;
}

function local_write($input) {
    $local = $input;
    $local['added'] = 'inside';
    return $local;
}

function copy_tree($input) {
    $out = [];
    foreach ($input as $key => $value) {
        if (is_array($value)) {
            $out[$key] = copy_tree($value);
        }
        if (is_string($value)) {
            $out[$key] = $value;
        }
    }
    return $out;
}

$ordered = ['02' => 'leading', '2' => 'two', 'name' => 'name'];
$ordered[2] = 'two-replaced';
$ordered['2'] = 'two-updated';
$ordered[5] = 'five';
$ordered['05'] = 'leading-five';
echo 'order=';
foreach ($ordered as $key => $value) {
    echo $key . '=' . $value . ';';
}
echo "\n";

$nulls = [];
$nulls['literal'] = null;
$nulls[7] = 'seven';
echo 'null=' . ($nulls['literal'] === null) . ':' . $nulls[7] . "\n";

$source = ['nested' => ['state' => 'original'], 'keep' => 'source'];
$copy = $source;
$nested = $copy['nested'];
$nested['state'] = 'changed';
$copy['nested'] = $nested;
echo 'copy=' . $source['nested']['state'] . ':' . $copy['nested']['state'] . ':' . $source['keep'] . "\n";

$self = ['first' => 'one', 'second' => 'two'];
$self['copy'] = $self;
echo 'self=' . $self['copy']['first'] . ':' . count($self['copy']) . ':' . count($self) . "\n";

$caller = ['keep' => 'caller'];
$local = local_write($caller);
echo 'local=' . $caller['keep'] . ':' . count($caller) . ':' . $local['added'] . ':' . count($local) . "\n";

$iterated = ['a' => 1, 'b' => 2];
foreach ($iterated as $key => $value) {
    // The added key must not become a further foreach visit.
    echo 'visit=' . $key . ':' . $value . ';';
    $iterated[$key] = $value + 10;
    $iterated['new'] = 100;
}
echo "\n";
echo 'foreach=' . $iterated['a'] . ':' . $iterated['b'] . ':' . $iterated['new'] . ':' . count($iterated) . "\n";

$tree = ['branch' => ['leaf' => 'L'], 'top' => 'T'];
$copied_tree = copy_tree($tree);
echo 'tree=' . $tree['branch']['leaf'] . ':' . $copied_tree['branch']['leaf'] . ':' . $copied_tree['top'] . "\n";

$markers = [];
$markers[marker_key('x')] = marker_value('one');
echo ':' . $markers['x'] . "\n";
