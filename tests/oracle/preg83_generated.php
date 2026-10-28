<?php
// Source-only B3 compiler fixture. The Go integration test appends its own
// observer so this file contains no PHP-side reporting or execution.

function preg83_basic() {
    $matches = null;
    $result = preg_match('/(a)(b)/s', 'zab', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_nomatch() {
    $matches = null;
    $result = preg_match('/z/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_named_unmatched() {
    $matches = null;
    $result = preg_match('/(?<letter>a)(b)?(c)?/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_negative_offset() {
    $matches = null;
    $result = preg_match('/a/s', 'xxa', $matches, PREG_OFFSET_CAPTURE, -1);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_middle_capture(string $subject, int $offset): array {
    $matches = null;
    $result = preg_match('/(a)(b)?(c)/s', $subject, $matches, PREG_OFFSET_CAPTURE, $offset);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_capture_copy() {
    $matches = null;
    $result = preg_match('/(?<letter>a)/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);
    $saved = $matches;
    $matches[0] = ['changed', 0];
    return ['saved' => $saved, 'changed' => $matches, 'result' => $result];
}

function preg83_binary_subject() {
    // The test turns the placeholder below into one literal NUL byte before both PHP and
    // GoHP parse the source; PHP single-quoted \x00 would not be a byte.
    $subject = 'left@NUL@right';
    $matches = null;
    $result = preg_match('/\\x00/s', $subject, $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_utf8_byte_offset() {
    $subject = '雪a';
    $matches = null;
    $result = preg_match('/a/s', $subject, $matches, PREG_OFFSET_CAPTURE, 3);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_lookahead_possessive() {
    $matches = null;
    $result = preg_match('/a(?=b)b*+/s', 'abbb', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

// Named property inputs are deliberately covered here rather than against a
// WordPress source file. The compiler admits only pure, named chains; the
// observer below reports scalars and arrays, never the generated objects.
class Preg83PropertyInput {
    public $subject = null;
    public $offset = null;
}

class Preg83PropertyHolder {
    public $input = null;
}

class Preg83PropertyReader {
    public $subject = null;
    public $offset = null;

    public function scan() {
        $matches = null;
        $result = preg_match('/(a)(b)/s', $this->subject, $matches, PREG_OFFSET_CAPTURE, $this->offset);
        return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error(), 'subject' => $this->subject, 'offset' => $this->offset];
    }
}

function preg83_property_this() {
    $reader = new Preg83PropertyReader();
    $reader->subject = 'zab';
    $reader->offset = 1;
    return $reader->scan();
}

function preg83_property_nested_utf8() {
    $input = new Preg83PropertyInput();
    $holder = new Preg83PropertyHolder();
    $holder->input = $input;
    $holder->input->subject = '雪a';
    $holder->input->offset = -1;
    $matches = null;
    $result = preg_match('/a/s', $holder->input->subject, $matches, PREG_OFFSET_CAPTURE, $holder->input->offset);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error(), 'subject' => $holder->input->subject, 'offset' => $holder->input->offset];
}

function preg83_property_mutation() {
    $input = new Preg83PropertyInput();
    $first_matches = null;
    $input->subject = 'zab';
    $input->offset = 1;
    $first = preg_match('/(a)(b)/s', $input->subject, $first_matches, PREG_OFFSET_CAPTURE, $input->offset);
    $second_matches = null;
    $input->subject = 'xxa';
    $input->offset = -1;
    $second = preg_match('/a/s', $input->subject, $second_matches, PREG_OFFSET_CAPTURE, $input->offset);
    return ['first' => $first, 'first_matches' => $first_matches, 'second' => $second, 'second_matches' => $second_matches, 'error' => preg_last_error(), 'subject' => $input->subject, 'offset' => $input->offset];
}

function preg83_property_capture_copy() {
    $input = new Preg83PropertyInput();
    $input->subject = 'a';
    $input->offset = 0;
    $matches = null;
    $result = preg_match('/(?<letter>a)/s', $input->subject, $matches, PREG_OFFSET_CAPTURE, $input->offset);
    $saved = $matches;
    $matches[0] = ['changed', 0];
    return ['saved' => $saved, 'changed' => $matches, 'result' => $result, 'error' => preg_last_error(), 'subject' => $input->subject, 'offset' => $input->offset];
}

function preg83_failure() {
    $subject = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!';
    $matches = null;
    $result = preg_match('/^(a+)+$/s', $subject, $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_read_error() {
    return preg_last_error();
}

function preg83_make_reader() {
    return fn() => preg_last_error();
}

function preg83_invoke_reader(callable $reader) {
    return $reader();
}

class ErrorProbe {
    public function read_error() {
        return preg_last_error();
    }
}

function preg83_shared_execution() {
    $failed = preg83_failure();
    $reader = preg83_make_reader();
    $from_arrow = preg83_invoke_reader($reader);
    $probe = new ErrorProbe();
    $from_method = $probe->read_error();
    // This successful named-property input call resets the same Execution's
    // measured error state after the backtrack failure.
    $reset = preg83_property_this();
    return [
        'failure' => $failed,
        'arrow_error' => $from_arrow,
        'method_error' => $from_method,
        'reset' => $reset,
        'final_error' => preg_last_error(),
    ];
}
