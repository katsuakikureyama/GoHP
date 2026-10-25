<?php
// Owned differential fixture: capture-free generated arrows and callable parameters.
function all_match(array $items, callable $callback): bool {
    foreach ($items as $key => $value) {
        if (!$callback($value, $key)) {
            return false;
        }
    }
    return true;
}

function forwarded_match(array $items, callable $callback): bool {
    return all_match($items, $callback);
}

function labelled_match($label, array $items, callable $callback): bool {
    echo '<' . $label . '>';
    return all_match($items, $callback);
}

function run_zero(callable $callback): bool {
    return $callback();
}

function delayed_call(callable $callback): bool {
    return $callback('later', 8);
}

function forwarded_call(callable $callback): bool {
    return delayed_call($callback);
}

function stop_marker($value, $key): bool {
    echo '<' . $key . '>';
    return $value !== 'stop';
}

function change_value(array $value): bool {
    $value['changed'] = 'yes';
    return isset($value['changed']);
}

function argument_marker($value) {
    echo '<arg:' . $value . '>';
    return $value;
}

$numeric = [-4 => 'first', 7 => 'last'];
$mixed = [-4 => 'first', 'name' => 'second'];
$key_is_numeric = fn($value, $key) => !is_string($key);
$numeric_result = all_match($numeric, $key_is_numeric);
$mixed_result = forwarded_match($mixed, $key_is_numeric);
$empty_result = all_match([], fn($value, $key) => stop_marker($value, $key));
$labelled_result = labelled_match('mixed', $numeric, fn($value, $key) => !is_string($key));
$short_result = all_match([-4 => 'go', 'middle' => 'stop', 3 => 'skip'], fn($value, $key) => stop_marker($value, $key));
$delayed_result = forwarded_call(fn($value, $key) => $value === 'later');
$repeated_first = all_match([1, 2], $key_is_numeric);
$repeated_second = all_match([3, 4], $key_is_numeric);
$zero_result = run_zero(fn() => true);
$nested_result = run_zero(fn() => run_zero(fn() => true));
$mutable = [['child' => 'original']];
$mutation_result = all_match($mutable, fn($value, $key) => change_value($value));
$original_child = $mutable[0];
$join = fn($left, $right) => $left . ':' . $right;
$ordered = $join(argument_marker('left'), argument_marker('right'));
$make = fn() => fn($value) => $value;
$returned = $make();
$returned_value = $returned('returned');

echo 'numeric:' . ($numeric_result ? 'true' : 'false') . "\n";
echo 'mixed:' . ($mixed_result ? 'true' : 'false') . "\n";
echo 'empty:' . ($empty_result ? 'true' : 'false') . "\n";
echo 'labelled:' . ($labelled_result ? 'true' : 'false') . "\n";
echo 'short:' . ($short_result ? 'true' : 'false') . "\n";
echo 'delayed:' . ($delayed_result ? 'true' : 'false') . "\n";
echo 'repeated:' . ($repeated_first ? 'true' : 'false') . ':' . ($repeated_second ? 'true' : 'false') . "\n";
echo 'zero:' . ($zero_result ? 'true' : 'false') . ':' . ($nested_result ? 'true' : 'false') . "\n";
echo 'mutation:' . ($mutation_result ? 'true' : 'false') . ':' . $mutable[0]['child'] . ':' . (isset($original_child['changed']) ? 'changed' : 'clean') . "\n";
echo 'ordered:' . $ordered . ':' . $returned_value . "\n";
