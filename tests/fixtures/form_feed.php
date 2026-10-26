<?php
// Owned differential fixture for PHP 8.3 form-feed escapes. A double-quoted
// \f is one byte; a single-quoted \f and double-quoted \\f are two literal
// bytes. The values are exercised through concatenation, array keys, function
// arguments/returns, and strtok's byte-delimiter set.

function bool_word($value) {
    if ($value === true) {
        return 'T';
    }
    return 'F';
}

function preserve_string($value) {
    return $value;
}

$form_feed = "\f";
$single_quoted = '\f';
$double_escaped = "\\f";
echo 'forms=' . strlen($form_feed) . ':' . strlen($single_quoted) . ':'
    . strlen($double_escaped) . ':' . bool_word($single_quoted === $double_escaped)
    . ':' . bool_word($form_feed === $single_quoted) . "\n";

$repeated = "\f\f";
$embedded = "A\fB\fC";
echo 'embedded=' . strlen($repeated) . ':' . strlen($embedded) . ':'
    . bool_word($repeated === $form_feed . $form_feed) . ':'
    . bool_word($embedded === 'A' . $form_feed . 'B' . $form_feed . 'C') . "\n";

$controls = "\n\t\f\r";
echo 'controls=' . strlen($controls) . ':'
    . bool_word($controls === "\n\t" . $form_feed . "\r") . "\n";

$by_escape = ["\f" => 'form-feed', '\f' => 'single-quoted', "\\f" => 'double-escaped'];
$returned = preserve_string($form_feed);
echo 'keys=' . $by_escape[$returned] . ':' . $by_escape[$single_quoted] . ':'
    . $by_escape[$double_escaped] . ':' . bool_word($returned === $form_feed) . "\n";

// Form feed is one delimiter byte alongside space, tab, carriage return, and
// newline. The letter f is ordinary token content and must not split a token.
$source = "one \t\f\r\nletterf\ffour";
$saved = $source;
$token = strtok($source, " \t\f\r\n");
$trace = '';
while ($token !== false) {
    $trace = $trace . '[' . $token . ']';
    $token = strtok(" \t\f\r\n");
}
echo 'strtok=' . $trace . ':' . bool_word($source === $saved) . "\n";
