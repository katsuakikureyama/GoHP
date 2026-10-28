<?php
// Owned differential fixture for the deliberately narrow list() lane. List
// targets are bare locals; numeric keys select by key rather than insertion
// order, and every right-hand side is observed exactly once.

function once($value) {
    echo '<once>';
    return $value;
}

function unpack_function($items) {
    list($first, $second) = $items;
    return $first . ':' . $second;
}

class ListAssignmentMethod {
    public function unpack_method($items) {
        list($first, $second) = $items;
        return $first . ':' . $second;
    }
}

$reordered = [
    4 => 'four',
    'extra' => 'ignored',
    1 => 'one',
    3 => 'three',
    0 => 'zero',
    2 => 'two',
];
list($zero, $one, $two, $three, $four) = $reordered;
echo 'five=' . $zero . ':' . $one . ':' . $two . ':' . $three . ':' . $four . "\n";

list($outer, $inner_pair) = ['outside', ['inside-one', 'inside-two']];
echo 'nested=' . $outer . ':' . $inner_pair[0] . ':' . $inner_pair[1] . "\n";

list($nullable, $word) = [null, 'word'];
echo 'null=' . (is_null($nullable) ? 'yes' : 'no') . ':' . $word . "\n";

list($duplicate, $duplicate) = ['first', 'second'];
echo 'duplicate=' . $duplicate . "\n";

$left = 'left';
$right = 'right';
list($left, $right) = [$right, $left];
echo 'swap=' . $left . ':' . $right . "\n";

$source = [['child' => 'source'], 'tail'];
$alias = $source;
list($copy, $tail) = $alias;
$copy['child'] = 'changed';
echo 'copy=' . $source[0]['child'] . ':' . $alias[0]['child'] . ':' . $copy['child'] . ':' . $tail . "\n";

// Replace the source variable itself only after its entire RHS is captured.
list($alias, $aliased_tail) = $alias;
echo 'source-alias=' . $alias['child'] . ':' . $aliased_tail . "\n";
list($capture, $position) = ['match', 7];
echo 'capture=' . $capture . ':' . $position . "\n";

$iteration = 0;
$loop_value = 'old';
while ($iteration < 2) {
    list($loop_value, $iteration) = ['new', $iteration + 1];
}
if (true) { list($loop_value) = ['branch']; }
echo 'while-if=' . $loop_value . ':' . $iteration . "\n";

list($marked_first, $marked_second) = once(['marker-first', 'marker-second']);
echo 'once=' . $marked_first . ':' . $marked_second . "\n";

echo 'function=' . unpack_function(['function-first', 'function-second']) . "\n";
$method = new ListAssignmentMethod();
echo 'method=' . $method->unpack_method(['method-first', 'method-second']) . "\n";

$rows = [[0, 'first'], [1, 'second'], [2, 'third']];
$loop = '';
foreach ($rows as $row) {
    list($number, $label) = $row;
    $loop = $loop . $number . ':' . $label . ';';
}
echo 'loop=' . $loop . "\n";
