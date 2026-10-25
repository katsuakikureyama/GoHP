<?php
// Strict identity with a literal null is a null predicate over the other
// operand. This fixture keeps general array/object identity out of scope.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function dynamic_value($value) {
    return $value;
}

function marker($name, $value) {
    echo '<' . $name . '>';
    return $value;
}

class Probe {}

$known_array = [];
$dynamic_array = dynamic_value($known_array);
$object = new Probe();
$callable = fn($value) => $value;

$left_equal = null === $known_array;
$right_equal = $known_array === null;
$left_not_equal = null !== $dynamic_array;
$right_not_equal = $dynamic_array !== null;
$null_equal = null === null;
$null_not_equal = null !== null;
$falsy = [false, 0, ''];
$object_equal = null === $object;
$object_not_equal = $object !== null;
$callable_equal = null === $callable;
$callable_not_equal = $callable !== null;
$once_left = null === marker('left', $known_array);
$once_right = marker('right', $known_array) !== null;
$short_and = false && (null === marker('and-skip', $known_array));
$short_or = true || (marker('or-skip', $known_array) !== null);

echo 'arrays=' . bool_word($left_equal) . ':' . bool_word($right_equal) . ':' . bool_word($left_not_equal) . ':' . bool_word($right_not_equal) . "\n";
echo 'null=' . bool_word($null_equal) . ':' . bool_word($null_not_equal) . "\n";
echo 'falsy=' . bool_word(null === $falsy[0]) . ':' . bool_word($falsy[1] === null) . ':' . bool_word(null !== $falsy[2]) . "\n";
echo 'other=' . bool_word($object_equal) . ':' . bool_word($object_not_equal) . ':' . bool_word($callable_equal) . ':' . bool_word($callable_not_equal) . "\n";
echo 'once=' . bool_word($once_left) . ':' . bool_word($once_right) . "\n";
echo 'short=' . bool_word($short_and) . ':' . bool_word($short_or) . "\n";
