<?php
function marker($name, $value) {
    echo '<' . $name . '>';
    return $value;
}

function compact_chain($value) {
    if (marker('c1', $value === 'first')) { return marker('c-first', 'A'); }
    elseif (marker('c2', $value === 'second')) { return marker('c-second', 'B'); }
    elseif (marker('c3', $value === 'third')) { return marker('c-third', 'C'); }
    else { return marker('c-fallback', 'D'); }
}

function spaced_chain($value) {
    if (marker('s1', $value === 'first')) { return marker('s-first', 'A'); }
    else if (marker('s2', $value === 'second')) { return marker('s-second', 'B'); }
    else if (marker('s3', $value === 'third')) { return marker('s-third', 'C'); }
    else { return marker('s-fallback', 'D'); }
}

function mixed_chain($value) {
    if (marker('m1', $value === 'first')) { return marker('m-first', 'A'); }
    elseif (marker('m2', $value === 'second')) { return marker('m-second', 'B'); }
    else if (marker('m3', $value === 'third')) { return marker('m-third', 'C'); }
    elseif (marker('m4', $value === 'fourth')) { return marker('m-fourth', 'D'); }
    else { return marker('m-fallback', 'E'); }
}

function truthy_chain($value) {
    if (false) { return 'never'; }
    elseif ($value) { return 'truthy'; }
    else { return 'falsey'; }
}

function no_final_else($value) {
    $state = 'none';
    if ($value === 'one') { $state = 'one'; }
    elseif ($value === 'two') { $state = 'two'; }
    return $state;
}

function empty_selected() {
    $state = 'keep';
    if (false) { $state = 'first'; }
    elseif (marker('empty', true)) {
    }
    elseif (marker('late', true)) { $state = 'late'; }
    else { $state = 'fallback'; }
    return $state;
}

function nested_chain($outer, $inner) {
    if ($outer) {
        if ($inner) { return 'inner-then'; }
        else { return 'inner-else'; }
    } elseif ($inner) { return 'outer-elseif'; }
    else { return 'outer-else'; }
}

function update_arrays($value) {
    $original = ['base'];
    $chosen = $original;
    $state = 'start';
    if ($value === 'append') { $local = 'A'; $state = 'append'; $chosen[] = $local; }
    elseif ($value === 'write') { $local = 'W'; $state = 'write'; $chosen['branch'] = $local; }
    else { $local = 'F'; $state = 'fallback'; $chosen[] = $local; }
    return $state . ':' . count($original) . ':' . count($chosen) . ':' . $original[0];
}

$empty = [];
$items = [0, 'value'];

echo 'compact=' . compact_chain('first') . ':' . compact_chain('second') . ':' . compact_chain('third') . ':' . compact_chain('other') . "\n";
echo 'spaced=' . spaced_chain('first') . ':' . spaced_chain('second') . ':' . spaced_chain('third') . ':' . spaced_chain('other') . "\n";
echo 'mixed=' . mixed_chain('first') . ':' . mixed_chain('second') . ':' . mixed_chain('third') . ':' . mixed_chain('fourth') . ':' . mixed_chain('other') . "\n";
echo 'truthy=' . truthy_chain(null) . ':' . truthy_chain(false) . ':' . truthy_chain(0) . ':' . truthy_chain('') . ':' . truthy_chain('0') . ':' . truthy_chain($empty) . ':' . truthy_chain(-7) . ':' . truthy_chain('value') . ':' . truthy_chain($items) . "\n";
echo 'none=' . no_final_else('one') . ':' . no_final_else('two') . ':' . no_final_else('other') . "\n";
echo 'empty=' . empty_selected() . "\n";
echo 'nested=' . nested_chain(true, true) . ':' . nested_chain(true, false) . ':' . nested_chain(false, true) . ':' . nested_chain(false, false) . "\n";
echo 'arrays=' . update_arrays('append') . ':' . update_arrays('write') . ':' . update_arrays('other') . "\n";
$changed = 0;
if (false) {} elseif (true) { $changed = ['late']; }
echo 'changed-type=' . count($changed) . ':' . $changed[0] . "\n";
$sibling = [];
if (false) { $sibling = 'earlier-only-type'; }
elseif (count($sibling) === 0) { $sibling[] = 'independent'; }
echo 'sibling-type=' . count($sibling) . ':' . $sibling[0] . "\n";
