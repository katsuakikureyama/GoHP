<?php
// Owned differential fixture: gettype() returns PHP's native type labels for
// the bounded GoHP value domain. The labelled output makes a numeric string
// distinct from an integer and keeps object/callable classification observable.

function describe_type(string $type) {
    return strtolower($type) . ':' . strlen($type);
}

function typed_kind(int $number) {
    return gettype($number);
}

function marked_value($value) {
    echo '<marked>';
    return $value;
}

class TypeBox {
    public $value = 'property';
}

$integer = 7;
$numeric_string = '7';
$nothing = null;
$false_value = false;
$empty_string = '';
$nested = ['inside' => ['number' => 9, 'text' => 'nested']];
$box = new TypeBox();

// This arrow is intentionally never invoked. Its own gettype expression must
// still compile, while the arrow value itself has PHP's object label.
$arrow = fn($value) => gettype($value);

$integer_type = GeTtYpE($integer);
$numeric_string_type = gettype($numeric_string);
$null_type = gettype($nothing);
$false_type = gettype($false_value);
$empty_string_type = gettype($empty_string);
$array_type = gettype($nested);
$nested_integer_type = gettype($nested['inside']['number']);
$nested_string_type = gettype($nested['inside']['text']);
$property_type = gettype($box->value);
$arrow_type = gettype($arrow);
$object_type = gettype($box);
$marked_type = gettype(marked_value(11));

// A discarded direct call remains an expression statement: it must evaluate
// once, including the visible marker in its argument.
gettype(marked_value('discarded'));

// strtok enables request-local execution plumbing; gettype observes its scalar
// result without making any additional tokenizer-state claim.
$token = strtok('first,second', ',');
$token_type = gettype($token);

echo 'scalars=' . $integer_type . ':' . $numeric_string_type . ':' . $null_type
    . ':' . $false_type . ':' . $empty_string_type . "\n";
echo 'compound=' . $array_type . ':' . $nested_integer_type . ':' . $nested_string_type
    . ':' . $property_type . ':' . $object_type . ':' . $arrow_type . "\n";
echo 'strings=' . describe_type($integer_type) . ':' . describe_type($numeric_string_type)
    . ':' . describe_type($array_type) . "\n";
echo 'trace=' . $marked_type . "\n";
echo 'token=' . $token_type . ':' . $token . "\n";
echo 'edges=' . gettype([]) . ':' . gettype(true) . ':' . gettype(0) . ':' . gettype('0') . ':' . typed_kind(-1) . "\n";
