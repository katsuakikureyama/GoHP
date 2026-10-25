<?php
// Original composition fixture: predicates choose the one safe transform.
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

function project($unknown) {
    return is_array($unknown) ? array_values($unknown) : (is_string($unknown) ? [stripslashes($unknown)] : []);
}

function marker($name, $value) {
    echo '<' . $name . '>';
    return $value;
}

function marked_project($unknown) {
    return is_array(marker('condition', $unknown)) ? marked_array($unknown) : (is_string($unknown) ? marked_string($unknown) : marked_empty());
}

function marked_array($value) {
    echo '[A]';
    return array_values($value);
}

function marked_string($value) {
    echo '[S]';
    return [stripslashes($value)];
}

function marked_empty() {
    echo '[E]';
    return [];
}

class Token {
}

$empty = [];
$list = ['first', 'second'];
$assoc = ['left' => 'one', 'right' => 'two'];
$nested = [['inner' => 'value']];
$escaped = 'O\\\'Brien';
$none = null;
$number = 42;
$object = new Token();

$empty_out = project(dynamic_value($empty));
$list_out = project(dynamic_value($list));
$assoc_out = project(dynamic_value($assoc));
$nested_out = project(dynamic_value($nested));
$string_out = project(dynamic_value($escaped));
$none_out = project(dynamic_value($none));
$number_out = project(dynamic_value($number));
$object_out = project(dynamic_value($object));

echo 'shape=' . count($empty_out) . ':' . $list_out[0] . ':' . $assoc_out[0] . ':' . $nested_out[0]['inner'] . ':' . $string_out[0] . ':' . count($none_out) . ':' . count($number_out) . ':' . count($object_out) . "\n";
echo 'strict=' . bool_word(is_array($empty)) . ':' . bool_word(is_array($object)) . ':' . bool_word(is_bool(is_array($nested))) . ':' . bool_word(is_array($nested) === true) . ':' . bool_word(is_array($escaped) === false) . "\n";
echo 'arms=' . count(marked_project(marker('array', $assoc))) . ':' . marked_project(marker('string', $escaped))[0] . ':' . count(marked_project(marker('null', $none))) . "\n";
echo 'original=' . $list[0] . ':' . $assoc['left'] . ':' . $nested[0]['inner'] . ':' . $escaped . "\n";
