<?php
// Owned differential fixture: this is the bounded, byte-delimiter strtok
// contract. It deliberately uses only initialized continuations; PHP 8.3
// warns on continuation after a scan was cleared by skipping only delimiters.

function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

// This nested helper continues the same request-local tokenizer. Its argument
// is a delimiter byte-set, not a delimiter substring.
function next_token($delimiters) {
    return strtok($delimiters);
}

function replace_scan() {
    return strtok('nested,new', ',');
}

function mapped_token($ignored) {
    return strtok(',;');
}

class ConstructorStep {
    public $token;
    public function __construct() {
        $this->token = strtok(',');
    }
}

class TokenStepper {
    public function next($delimiters) {
        return strtok($delimiters);
    }
}

// A two-argument call initializes/replaces the scan. Changing the delimiter
// set on a one-argument call applies to the remaining bytes.
$changed_source = 'red,blue;green';
$changed_saved = $changed_source;
$changed_first = strtok($changed_source, ',');
$changed_second = next_token(';');
$changed_third = strtok(';');
echo 'changed=' . $changed_first . ':' . $changed_second . ':' . $changed_third
    . ':' . bool_word($changed_source === $changed_saved) . "\n";

// Comma, semicolon, and vertical bar are three independent separator bytes.
// Adjacent and terminal separators do not produce empty tokens.
$multi_source = 'one,,two;three|four|||';
$multi_saved = $multi_source;
$multi = strtok($multi_source, ',;|');
$multi_trace = '';
while ($multi !== false) {
    $multi_trace = $multi_trace . '[' . $multi . ']';
    $multi = strtok(',;|');
}
echo 'multi=' . $multi_trace . ':' . bool_word($multi_source === $multi_saved) . "\n";

// An input with no token returns false on the initializing call. Do not make
// a further continuation call after either exhausted scan.
$empty_input = strtok('', ',;');
$only_separators = strtok(" ,;\t", " ,;\t");
echo 'empty=' . bool_word($empty_input === false) . ':' . bool_word($only_separators === false) . "\n";
$no_delimiter = strtok(' a,b ', '');
$no_more = strtok('');
$binary_token = strtok('é,c', ',');
echo 'bytes=' . $no_delimiter . ':' . bool_word($no_more === false) . ':' . strlen($binary_token) . "\n";
$discarded = strtok('outer,tail', ',');
$nested_reset = replace_scan();
echo 'nested-reset=' . $discarded . ':' . $nested_reset . ':' . strtok(',') . "\n";

// Resetting with a new input discards the previous cursor rather than reviving
// it. This also demonstrates that the literal string token "0" is not false.
$reset_first = strtok('old,tail', ',');
$reset_second = strtok('zero|0|last', '|');
$reset_third = strtok('|');
$reset_fourth = strtok('|');
echo 'reset=' . $reset_first . ':' . $reset_second . ':' . $reset_third . ':' . $reset_fourth
    . ':' . bool_word($reset_third === '0') . "\n";

// A normal untyped method and a capture-free arrow both continue the exact
// same execution state as their caller.
$stepper = new TokenStepper();
$method_first = strtok('method:next:tail', ':');
$method_second = $stepper->next(':');
$method_third = strtok(':');

$arrow_next = fn($ignored) => strtok('/');
$arrow_first = strtok('arrow/next/tail', '/');
$arrow_second = $arrow_next('ignored');
$arrow_third = strtok('/');
echo 'shared=' . $method_first . ':' . $method_second . ':' . $method_third
    . ':' . $arrow_first . ':' . $arrow_second . ':' . $arrow_third . "\n";
$constructor_first = strtok('before,constructor,after', ',');
$constructor = new ConstructorStep();
echo 'constructor=' . $constructor_first . ':' . $constructor->token . ':' . strtok(',') . "\n";

// Literal named array_map callbacks use that same state; the exhausted third
// callback value is false and is retained as false in the result array.
$map_first = strtok('map-a,map-b;map-c', ',;');
$mapped = array_map('mapped_token', ['one', 'two', 'three']);
echo 'map=' . $map_first . ':' . $mapped[0] . ':' . $mapped[1]
    . ':' . bool_word($mapped[2] === false) . "\n";

// An owned whitespace-token selection helper models only the prefix decision:
// it is not a WordPress or HTML-processor execution claim. Leading, repeated,
// and trailing whitespace are separator bytes, and break keeps the first
// matching token while leaving its string input unchanged.
$whitespace_source = " \tplain\nwp-elements-card   later ";
$whitespace_saved = $whitespace_source;
$token = strtok($whitespace_source, " \t\n");
$selected = '';
while ($token !== false) {
    if (str_starts_with($token, 'wp-elements-')) {
        $selected = $token;
        break;
    }
    $token = strtok(" \t\n");
}
echo 'prefix=' . $selected . ':' . bool_word($whitespace_source === $whitespace_saved) . "\n";
