<?php
// Owned differential fixture: trailing null defaults are limited to public,
// untyped constructors. Omitted and explicit null arguments both reach the
// constructor as null; zero remains distinct through null coalescing.
function constructor_marker($name, $value) {
    echo '<arg:' . $name . '>';
    return $value;
}

class ThreeDefaults {
    public $first;
    public $second;
    public $third;

    public function __construct($first, $second = null, $third = null) {
        $this->first = $first;
        $this->second = $second ?? 20;
        $this->third = $third ?? 30;
    }
}

class FourDefaults {
    public $first;
    public $second;
    public $third;
    public $fourth;

    public function __construct($first, $second, $third = null, $fourth = null) {
        $this->first = $first;
        $this->second = $second;
        $this->third = $third ?? 30;
        $this->fourth = $fourth ?? 40;
    }
}

class FiveDefaults {
    public $first;
    public $second;
    public $third;
    public $fourth;
    public $fifth;

    public function __construct($first, $second, $third, $fourth = null, $fifth = null) {
        $this->first = $first;
        $this->second = $second;
        $this->third = $third;
        $this->fourth = $fourth ?? 40;
        $this->fifth = $fifth ?? 50;
    }
}

class AllOptional {
    public $first;
    public $second;

    public function __CONSTRUCT($first = null, $second = null) {
        $this->first = $first ?? 4;
        $this->second = $second ?? 8;
    }
}

class OrderedDefaults {
    public $first;
    public $second;

    public function __construct($first, $second = null) {
        echo '<body>';
        $this->first = $first;
        $this->second = $second ?? 'default';
    }
}

class ConstructorTokenizer {
    public $token;

    public function __construct($unused = null) {
        $this->token = strtok(',');
    }
}

$three_omitted = new ThreeDefaults('one');
$three_explicit = new ThreeDefaults('two', null, 0);
$three_values = new ThreeDefaults('three', 2, 3);
$four_omitted = new FourDefaults('one', 'two');
$four_explicit = new FourDefaults('three', 'four', null, 0);
$five_omitted = new FiveDefaults('one', 'two', 'three');
$five_explicit = new FiveDefaults('four', 'five', 'six', null, 0);
$optional_omitted = new AllOptional();
$optional_explicit = new AllOptional(null, 0);

$ordered = new OrderedDefaults(
    constructor_marker('first', 'A'),
    constructor_marker('second', 'B')
);

$fresh_first = new FourDefaults('first', 'shared');
$fresh_second = new FourDefaults('second', 'other');
$fresh_first->first = 'changed';

$token_first = strtok('caller,constructor,after', ',');
$tokenized = new ConstructorTokenizer();
$token_after = strtok(',');

echo "\nthree=" . $three_omitted->first . ':' . $three_omitted->second . ':' . $three_omitted->third
    . ':' . $three_explicit->first . ':' . $three_explicit->second . ':' . $three_explicit->third
    . ':' . $three_values->first . ':' . $three_values->second . ':' . $three_values->third . "\n";
echo 'four=' . $four_omitted->first . ':' . $four_omitted->second . ':' . $four_omitted->third . ':' . $four_omitted->fourth
    . ':' . $four_explicit->first . ':' . $four_explicit->second . ':' . $four_explicit->third . ':' . $four_explicit->fourth . "\n";
echo 'five=' . $five_omitted->first . ':' . $five_omitted->second . ':' . $five_omitted->third . ':' . $five_omitted->fourth . ':' . $five_omitted->fifth
    . ':' . $five_explicit->first . ':' . $five_explicit->second . ':' . $five_explicit->third . ':' . $five_explicit->fourth . ':' . $five_explicit->fifth . "\n";
echo 'all-optional=' . $optional_omitted->first . ':' . $optional_omitted->second
    . ':' . $optional_explicit->first . ':' . $optional_explicit->second . "\n";
echo 'ordered=' . $ordered->first . ':' . $ordered->second . "\n";
echo 'fresh=' . $fresh_first->first . ':' . $fresh_second->first . ':' . $fresh_second->second . "\n";
echo 'shared=' . $token_first . ':' . $tokenized->token . ':' . $token_after . "\n";
